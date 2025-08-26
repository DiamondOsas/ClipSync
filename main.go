package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"clipsync/modules"
)

var (
	serviceName = "ClipSync"
	serviceType = "_clipsync._tcp"
	port        = flag.Int("port", 50123, "Port to listen on")
	debug       = flag.Bool("debug", false, "Enable debug logging")
)

// App represents the main application state
type App struct {
	th               *material.Theme
	clipboardHistory *ClipboardHistory
	connectedDevices *ConnectedDevices
	
	peers      map[string]*modules.Peer
	connMutex  sync.RWMutex
	statusLog  []string
	statusMu   sync.Mutex
	
	currentClipEditor  widget.Editor
	previousClipEditor widget.Editor
	statusText         string
	deviceListState    widget.List
	
	sd     *modules.ServiceDiscovery
	ctx    context.Context
	cancel context.CancelFunc
	
	peerID           string
	tlsConfig        *tls.Config
	lastClipboardUpdate time.Time
}

// ClipboardHistory stores clipboard history
type ClipboardHistory struct {
	Current  string
	Previous string
	mu       sync.RWMutex
}

func (ch *ClipboardHistory) SetCurrent(content string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.Previous = ch.Current
	ch.Current = content
}

func (ch *ClipboardHistory) GetCurrent() string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.Current
}

func (ch *ClipboardHistory) GetPrevious() string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.Previous
}

// ConnectedDevices stores connected devices
type ConnectedDevices struct {
	devices []*modules.ServiceRecord
	mu      sync.RWMutex
}

func (cd *ConnectedDevices) SetDevices(devices []*modules.ServiceRecord) {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	cd.devices = devices
}

func (cd *ConnectedDevices) GetDevices() []*modules.ServiceRecord {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	devicesCopy := make([]*modules.ServiceRecord, len(cd.devices))
	copy(devicesCopy, cd.devices)
	return devicesCopy
}

func (cd *ConnectedDevices) AddDevice(device *modules.ServiceRecord) {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	
	for i, existing := range cd.devices {
		if len(existing.IPs) > 0 && len(device.IPs) > 0 && existing.IPs[0].Equal(device.IPs[0]) {
			cd.devices[i] = device
			return
		}
	}
	cd.devices = append(cd.devices, device)
}

func (cd *ConnectedDevices) Count() int {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	return len(cd.devices)
}

func NewApp() *App {
	ctx, cancel := context.WithCancel(context.Background())
	
	hostname, _ := os.Hostname()
	peerID := fmt.Sprintf("%s-%d", hostname, time.Now().UnixNano())
	
	cert, err := modules.GenerateSelfSignedCert()
	if err != nil {
		log.Printf("Error generating TLS certificate: %v", err)
	}
	
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}
	
	app := &App{
		th:               material.NewTheme(),
		clipboardHistory: &ClipboardHistory{},
		connectedDevices: &ConnectedDevices{},
		peers:            make(map[string]*modules.Peer),
		statusLog:        make([]string, 0, 50),
		sd:               &modules.ServiceDiscovery{},
		ctx:              ctx,
		cancel:           cancel,
		peerID:           peerID,
		tlsConfig:        tlsConfig,
		lastClipboardUpdate: time.Now(),
	}
	
	app.currentClipEditor.Submit = false
	app.currentClipEditor.SingleLine = false
	app.previousClipEditor.Submit = false
	app.previousClipEditor.SingleLine = false
	app.previousClipEditor.ReadOnly = true
	app.deviceListState.Axis = layout.Vertical
	
	return app
}

func (a *App) startServices() {
	flag.Parse()
	
	if *debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("=== ClipSync Debug Mode ===")
		log.Printf("Peer ID: %s", a.peerID)
	} else {
		log.SetOutput(os.Stdout)
	}
	
	// Start service discovery
	err := a.sd.StartServiceDiscovery(serviceName, serviceType, *port)
	if err != nil {
		a.updateStatus(fmt.Sprintf("Discovery error: %v", err))
		log.Printf("Error starting service discovery: %v", err)
	} else {
		a.updateStatus("Service discovery started")
		log.Printf("Service discovery started on port %d", *port)
	}

	// Start TLS listener
	go func() {
		log.Printf("Starting TLS listener on port %d", *port)
		err := modules.ListenForDataTLS(*port, a.tlsConfig, func(peer *modules.Peer, data []byte) {
			if len(data) == 0 {
				return
			}
			
			text := string(data)
			log.Printf("Received data from %s: %q", peer.ID, text)
			
			// Prevent clipboard loops
			a.connMutex.Lock()
			recentUpdate := time.Since(a.lastClipboardUpdate) < 500*time.Millisecond
			if recentUpdate {
				a.connMutex.Unlock()
				log.Printf("Skipping clipboard update to prevent loop")
				return
			}
			a.lastClipboardUpdate = time.Now()
			a.connMutex.Unlock()
			
			if err := modules.WriteClipboard(text); err != nil {
				a.updateStatus(fmt.Sprintf("Clipboard error: %v", err))
			} else {
				a.currentClipEditor.SetText(text)
				a.clipboardHistory.SetCurrent(text)
				a.previousClipEditor.SetText(a.clipboardHistory.GetPrevious())
				a.updateStatus(fmt.Sprintf("Received from %s", peer.ID))
			}
		})
		if err != nil {
			a.updateStatus(fmt.Sprintf("Listen error: %v", err))
			log.Printf("Error starting listener: %v", err)
		}
	}()

	// Start discovery loop
	go func() {
		discoveryTicker := time.NewTicker(3 * time.Second)
		reconnectTicker := time.NewTicker(10 * time.Second)
		defer discoveryTicker.Stop()
		defer reconnectTicker.Stop()
		
		a.discoverAndSyncDevices() // Initial discovery
		
		for {
			select {
			case <-a.ctx.Done():
				return
			case <-discoveryTicker.C:
				a.discoverAndSyncDevices()
			case <-reconnectTicker.C:
				a.reconnectToPeers()
			}
		}
	}()

	// Start clipboard monitor
	go func() {
		var lastContent string
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-a.ctx.Done():
				return
			case <-ticker.C:
				content, err := modules.ReadClipboard()
				if err != nil || content == "" || content == lastContent {
					continue
				}

				lastContent = content
				a.currentClipEditor.SetText(content)
				a.clipboardHistory.SetCurrent(content)
				a.previousClipEditor.SetText(a.clipboardHistory.GetPrevious())
				a.updateStatus("Clipboard updated")

				a.connMutex.Lock()
				a.lastClipboardUpdate = time.Now()
				a.connMutex.Unlock()

				a.sendClipboardToPeers([]byte(content))
			}
		}
	}()
}

func (a *App) discoverAndSyncDevices() {
	services, err := modules.DiscoverServices(serviceType, 3)
	if err != nil {
		a.updateStatus(fmt.Sprintf("Discovery error: %v", err))
		return
	}

	localIPs, err := getAllLocalIPs()
	if err != nil {
		log.Printf("Error getting local IPs: %v", err)
		return
	}

	if *debug {
		log.Printf("Found %d services, filtering local IPs: %v", len(services), localIPs)
	}

	validDevices := make([]*modules.ServiceRecord, 0)
	
	for _, service := range services {
		isSelf := false
		
		for _, serviceIP := range service.IPs {
			for _, localIP := range localIPs {
				if serviceIP.Equal(localIP) && service.Port == *port {
					isSelf = true
					break
				}
			}
			if isSelf {
				break
			}
		}
		
		if !isSelf && len(service.IPs) > 0 {
			validDevices = append(validDevices, service)
			a.connectedDevices.AddDevice(service)
			
			ip := service.IPs[0].String()
			
			a.connMutex.RLock()
			alreadyConnected := false
			for _, peer := range a.peers {
				peerIP, _, err := net.SplitHostPort(peer.Address)
				if err != nil {
					// If we can't parse the address, compare directly
					if peer.Address == ip || strings.HasPrefix(peer.Address, ip+":") {
						alreadyConnected = true
						break
					}
				} else if peerIP == ip {
					alreadyConnected = true
					break
				}
			}
			
			// Log connection status for debugging
			if *debug {
				if alreadyConnected {
					log.Printf("Already connected to %s, skipping connection attempt", ip)
				} else {
					log.Printf("Not connected to %s, will attempt connection", ip)
				}
			}
			a.connMutex.RUnlock()
			
			if !alreadyConnected {
				go func(ipAddr string, name string) {
					log.Printf("Connecting to %s (%s)", name, ipAddr)
					peer, err := modules.ConnectToPeerTLS(ipAddr, *port, a.tlsConfig)
					if err != nil {
						log.Printf("Connection failed to %s: %v", ipAddr, err)
						a.updateStatus(fmt.Sprintf("Connection failed to %s", ipAddr))
					} else {
						a.connMutex.Lock()
						a.peers[peer.ID] = peer
						a.connMutex.Unlock()
						log.Printf("Connected to %s (%s)", name, ipAddr)
						a.updateStatus(fmt.Sprintf("Connected to %s", ipAddr))
					}
				}(ip, service.Name)
			}
		}
	}

	a.connectedDevices.SetDevices(validDevices)
	count := len(validDevices)
	if count > 0 {
		a.updateStatus(fmt.Sprintf("Discovered %d device(s)", count))
	} else {
		a.updateStatus("No devices found")
	}
	
	// Log current connections for debugging
	if *debug {
		a.connMutex.RLock()
		log.Printf("Current connections: %d peers", len(a.peers))
		for id, peer := range a.peers {
			log.Printf("  Peer %s: %s", id, peer.Address)
		}
		a.connMutex.RUnlock()
	}
}

func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(2)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		border := clip.Rect{Max: gtx.Constraints.Max}.Op()
		paint.FillShape(gtx.Ops, color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}, border)
		
		return layout.UniformInset(unit.Dp(2)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(0.7, func(gtx layout.Context) layout.Dimensions {
					return a.layoutClipboardPanel(gtx)
				}),
				layout.Flexed(0.3, func(gtx layout.Context) layout.Dimensions {
					return a.layoutDevicePanel(gtx)
				}),
			)
		})
	})
}

func (a *App) layoutClipboardPanel(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, color.NRGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xFF}, 
			clip.Rect{Max: gtx.Constraints.Min}.Op())
		
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := material.H4(a.th, "ClipSync")
				title.Color = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
				title.Alignment = text.Middle
				return layout.Inset{Bottom: unit.Dp(16)}.Layout(gtx, title.Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body1(a.th, "Current Clipboard:")
						label.Color = color.NRGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF}
						return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, label.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.Y = gtx.Dp(120)
						gtx.Constraints.Max.Y = gtx.Dp(120)
						editor := material.Editor(a.th, &a.currentClipEditor, "Current clipboard...")
						editor.Color = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
						return a.layoutEditorWithBackground(gtx, editor)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body1(a.th, "Previous Clipboard:")
						label.Color = color.NRGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF}
						return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, label.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.Y = gtx.Dp(80)
						gtx.Constraints.Max.Y = gtx.Dp(80)
						editor := material.Editor(a.th, &a.previousClipEditor, "Previous clipboard...")
						editor.Color = color.NRGBA{R: 0xC0, G: 0xC0, B: 0xC0, A: 0xFF}
						return a.layoutEditorWithBackground(gtx, editor)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutStatusSection(gtx)
			}),
		)
	})
}

func (a *App) layoutDevicePanel(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, color.NRGBA{R: 0x18, G: 0x18, B: 0x18, A: 0xFF}, 
			clip.Rect{Max: gtx.Constraints.Min}.Op())
		
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := material.H6(a.th, "Connected Devices")
				title.Color = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
				return layout.Inset{Bottom: unit.Dp(16)}.Layout(gtx, title.Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				count := a.connectedDevices.Count()
				label := material.Body2(a.th, fmt.Sprintf("%d device(s)", count))
				label.Color = color.NRGBA{R: 0xA0, G: 0xA0, B: 0xA0, A: 0xFF}
				return layout.Inset{Bottom: unit.Dp(12)}.Layout(gtx, label.Layout)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return a.layoutDeviceList(gtx)
			}),
		)
	})
}

func (a *App) layoutDeviceList(gtx layout.Context) layout.Dimensions {
	devices := a.connectedDevices.GetDevices()
	
	return material.List(a.th, &a.deviceListState).Layout(gtx, len(devices), func(gtx layout.Context, i int) layout.Dimensions {
		if i >= len(devices) {
			return layout.Dimensions{}
		}
		
		return a.layoutDeviceItem(gtx, devices[i])
	})
}

func (a *App) layoutDeviceItem(gtx layout.Context, device *modules.ServiceRecord) layout.Dimensions {
	return layout.Inset{
		Top: unit.Dp(4), Bottom: unit.Dp(4),
		Left: unit.Dp(8), Right: unit.Dp(8),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, color.NRGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xFF}, 
			clip.UniformRRect(image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Min.Y), 8).Op(gtx.Ops))
		
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					name := device.Name
					if name == "" {
						name = "Unknown Device"
					}
					label := material.Body1(a.th, name)
					label.Color = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
					return label.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					var ipStr string
					if len(device.IPs) > 0 {
						ipStr = device.IPs[0].String()
					} else {
						ipStr = "No IP"
					}
					label := material.Caption(a.th, ipStr)
					label.Color = color.NRGBA{R: 0xA0, G: 0xA0, B: 0xA0, A: 0xFF}
					return layout.Inset{Top: unit.Dp(2)}.Layout(gtx, label.Layout)
				}),
			)
		})
	})
}

func (a *App) layoutStatusSection(gtx layout.Context) layout.Dimensions {
	return layout.Inset{
		Top: unit.Dp(8), Bottom: unit.Dp(8),
		Left: unit.Dp(12), Right: unit.Dp(12),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, color.NRGBA{R: 0x28, G: 0x28, B: 0x28, A: 0xFF}, 
			clip.UniformRRect(image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Min.Y), 6).Op(gtx.Ops))
		
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body2(a.th, "Status:")
					label.Color = color.NRGBA{R: 0xB0, G: 0xB0, B: 0xB0, A: 0xFF}
					return label.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Caption(a.th, a.wrapText(a.statusText, 40))
					label.Color = color.NRGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF}
					return layout.Inset{Top: unit.Dp(4)}.Layout(gtx, label.Layout)
				}),
			)
		})
	})
}

func (a *App) layoutEditorWithBackground(gtx layout.Context, editor material.EditorStyle) layout.Dimensions {
	paint.FillShape(gtx.Ops, color.NRGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xFF}, 
		clip.UniformRRect(image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y), 4).Op(gtx.Ops))
	
	return layout.UniformInset(unit.Dp(8)).Layout(gtx, editor.Layout)
}

func (a *App) wrapText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	
	var result strings.Builder
	words := strings.Fields(text)
	currentLine := ""
	
	for _, word := range words {
		if len(currentLine)+len(word)+1 <= maxWidth {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		} else {
			if currentLine != "" {
				result.WriteString(currentLine + "\n")
			}
			currentLine = word
		}
	}
	
	if currentLine != "" {
		result.WriteString(currentLine)
	}
	
	return result.String()
}

func (a *App) shutdown() {
	a.cancel()
	
	if a.sd != nil {
		a.sd.StopServiceDiscovery()
	}
	
	a.connMutex.Lock()
	for id, peer := range a.peers {
		log.Printf("Closing connection to %s", id)
		peer.Close()
	}
	a.peers = make(map[string]*modules.Peer)
	a.connMutex.Unlock()
}

func (a *App) sendClipboardToPeers(data []byte) {
	a.connMutex.RLock()
	defer a.connMutex.RUnlock()
	
	for id, peer := range a.peers {
		if err := peer.SendClipboardData(data); err != nil {
			log.Printf("Failed to send to %s: %v", id, err)
		}
	}
}

func (a *App) reconnectToPeers() {
	// Get list of currently connected peers
	a.connMutex.RLock()
	peers := make([]*modules.Peer, 0, len(a.peers))
	for _, peer := range a.peers {
		peers = append(peers, peer)
	}
	a.connMutex.RUnlock()
	
	// Check each peer's connection
	for _, peer := range peers {
		if !peer.IsConnectionAlive() {
			// Connection is dead, remove it
			log.Printf("Peer %s connection lost, removing", peer.ID)
			a.connMutex.Lock()
			delete(a.peers, peer.ID)
			a.connMutex.Unlock()
			
			// Try to reconnect
			go func(peerAddr string) {
				// Extract IP from address (peer.Address is in format "ip:port")
				host, _, err := net.SplitHostPort(peerAddr)
				if err != nil {
					log.Printf("Failed to parse peer address %s: %v", peerAddr, err)
					return
				}
				
				log.Printf("Attempting to reconnect to %s", host)
				newPeer, err := modules.ConnectToPeerTLS(host, *port, a.tlsConfig)
				if err != nil {
					log.Printf("Failed to reconnect to %s: %v", host, err)
				} else {
					a.connMutex.Lock()
					a.peers[newPeer.ID] = newPeer
					a.connMutex.Unlock()
					log.Printf("Reconnected to %s", host)
					a.updateStatus(fmt.Sprintf("Reconnected to %s", host))
				}
			}(peer.Address)
		}
	}
}

func (a *App) updateStatus(msg string) {
	a.statusMu.Lock()
	defer a.statusMu.Unlock()
	
	timestamped := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)
	a.statusLog = append(a.statusLog, timestamped)
	
	if len(a.statusLog) > 10 {
		a.statusLog = a.statusLog[1:]
	}
	
	a.statusText = strings.Join(a.statusLog, "\n")
}

func getAllLocalIPs() ([]net.IP, error) {
	var ips []net.IP
	
	interfaces, err := net.Interfaces()
	if err != nil {
		return getDefaultIP()
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if !ip.IsLoopback() && ip.To4() != nil {
				ips = append(ips, ip)
			}
		}
	}

	if len(ips) == 0 {
		return getDefaultIP()
	}

	return ips, nil
}

func getDefaultIP() ([]net.IP, error) {
	hostname, _ := os.Hostname()
	addrs, _ := net.LookupHost(hostname)
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip != nil && !ip.IsLoopback() && ip.To4() != nil {
			return []net.IP{ip}, nil
		}
	}
	
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, fmt.Errorf("unable to determine local IP")
	}
	defer conn.Close()
	
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return []net.IP{localAddr.IP}, nil
}

func main() {
	flag.Parse()
	
	go func() {
		w := new(app.Window)
		w.Option(app.Title("ClipSync"))
		w.Option(app.Size(unit.Dp(800), unit.Dp(600)))

		myApp := NewApp()
		myApp.startServices()
		defer myApp.shutdown()

		var ops op.Ops
		for {
			switch e := w.Event().(type) {
			case app.DestroyEvent:
				return
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)
				myApp.Layout(gtx)
				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}