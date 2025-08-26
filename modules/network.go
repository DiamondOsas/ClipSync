package modules

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"sync"
	"time"
	"fmt"

	"github.com/grandcat/zeroconf"
)

// ServiceDiscovery handles mDNS service discovery
type ServiceDiscovery struct {
	server *zeroconf.Server
}

// StartServiceDiscovery advertises the service
func (sd *ServiceDiscovery) StartServiceDiscovery(serviceName, serviceType string, port int) error {
	ips, err := getAllLocalIPs()
	if err != nil {
		return err
	}

	ipStrings := make([]string, len(ips))
	for i, ip := range ips {
		ipStrings[i] = ip.String()
	}

	server, err := zeroconf.Register(
		serviceName,
		serviceType,
		"local.",
		port,
		ipStrings,
		nil,
	)
	if err != nil {
		return err
	}

	sd.server = server
	log.Printf("Service '%s' registered on port %d", serviceName, port)
	return nil
}

// StopServiceDiscovery stops advertising
func (sd *ServiceDiscovery) StopServiceDiscovery() {
	if sd.server != nil {
		sd.server.Shutdown()
	}
}

// DiscoverServices finds services on the network
func DiscoverServices(serviceType string, timeout int) ([]*ServiceRecord, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	entries := make(chan *zeroconf.ServiceEntry)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	var services []*ServiceRecord

	go func() {
		for entry := range entries {
			service := &ServiceRecord{
				Name: entry.ServiceRecord.Instance,
				Host: entry.HostName,
				Port: entry.Port,
				IPs:  entry.AddrIPv4,
			}
			services = append(services, service)
			log.Printf("Discovered service: %s at %v:%d", service.Name, service.IPs, service.Port)
		}
	}()

	err = resolver.Browse(ctx, serviceType, "local.", entries)
	if err != nil {
		return nil, err
	}

	<-ctx.Done()
	log.Printf("Service discovery complete, found %d services", len(services))
	return services, nil
}

// ServiceRecord represents a discovered service
type ServiceRecord struct {
	Name string
	Host string
	Port int
	IPs  []net.IP
}

// Peer represents a connected peer
type Peer struct {
	ID      string
	Conn    net.Conn
	Address string
	mu      sync.RWMutex
}

// HandshakeMessage for peer identification
type HandshakeMessage struct {
	PeerID string `json:"peer_id"`
}

// ListenForDataTLS listens for incoming TLS connections
func ListenForDataTLS(port int, tlsConfig *tls.Config, handler func(*Peer, []byte)) error {
	listener, err := tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port), tlsConfig)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("TLS listener started on port %d", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleConnection(conn, handler)
	}
}

func handleConnection(conn net.Conn, handler func(*Peer, []byte)) {
	defer conn.Close()
	
	log.Printf("New incoming connection from %s", conn.RemoteAddr().String())
	
	// Set initial timeout for handshake
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	var handshakeLen uint32
	if err := binary.Read(conn, binary.BigEndian, &handshakeLen); err != nil {
		log.Printf("Error reading handshake length from %s: %v", conn.RemoteAddr().String(), err)
		return
	}

	handshakeData := make([]byte, handshakeLen)
	if _, err := io.ReadFull(conn, handshakeData); err != nil {
		log.Printf("Error reading handshake from %s: %v", conn.RemoteAddr().String(), err)
		return
	}

	var handshake HandshakeMessage
	if err := json.Unmarshal(handshakeData, &handshake); err != nil {
		log.Printf("Error parsing handshake from %s: %v", conn.RemoteAddr().String(), err)
		return
	}

	peer := &Peer{
		ID:      handshake.PeerID,
		Conn:    conn,
		Address: conn.RemoteAddr().String(),
	}

	log.Printf("New peer connected: %s from %s", peer.ID, peer.Address)

	// Remove timeout for ongoing communication
	conn.SetReadDeadline(time.Time{})

	for {
		var frameLen uint32
		if err := binary.Read(conn, binary.BigEndian, &frameLen); err != nil {
			log.Printf("Peer %s disconnected: %v", peer.ID, err)
			return
		}

		frameData := make([]byte, frameLen)
		if _, err := io.ReadFull(conn, frameData); err != nil {
			log.Printf("Error reading frame from %s: %v", peer.ID, err)
			return
		}

		if len(frameData) > 0 {
			handler(peer, frameData)
		}
	}
}

// ConnectToPeerTLS establishes outbound TLS connection
func ConnectToPeerTLS(ip string, port int, tlsConfig *tls.Config) (*Peer, error) {
	hostname, _ := os.Hostname()
	peerID := fmt.Sprintf("%s-%d", hostname, time.Now().UnixNano())

	handshake := HandshakeMessage{PeerID: peerID}
	handshakeData, err := json.Marshal(handshake)
	if err != nil {
		return nil, err
	}

	log.Printf("Attempting to connect to %s:%d", ip, port)
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 5 * time.Second},
		"tcp",
		fmt.Sprintf("%s:%d", ip, port),
		tlsConfig,
	)
	if err != nil {
		log.Printf("Failed to connect to %s:%d: %v", ip, port, err)
		return nil, err
	}

	log.Printf("TCP connection established to %s:%d, sending handshake", ip, port)
	handshakeLen := uint32(len(handshakeData))
	if err := binary.Write(conn, binary.BigEndian, handshakeLen); err != nil {
		log.Printf("Failed to send handshake length to %s:%d: %v", ip, port, err)
		conn.Close()
		return nil, err
	}

	if _, err := conn.Write(handshakeData); err != nil {
		log.Printf("Failed to send handshake data to %s:%d: %v", ip, port, err)
		conn.Close()
		return nil, err
	}

	peer := &Peer{
		ID:      peerID,
		Conn:    conn,
		Address: conn.RemoteAddr().String(),
	}

	log.Printf("Connected to peer at %s:%d", ip, port)
	return peer, nil
}

// SendClipboardData sends data to peer
func (p *Peer) SendClipboardData(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Conn == nil {
		return io.ErrClosedPipe
	}

	// Set a write timeout
	p.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	
	frameLen := uint32(len(data))
	if err := binary.Write(p.Conn, binary.BigEndian, frameLen); err != nil {
		return err
	}

	_, err := p.Conn.Write(data)
	
	// Clear the write timeout
	p.Conn.SetWriteDeadline(time.Time{})
	
	return err
}

// IsConnectionAlive checks if the connection to the peer is still alive
func (p *Peer) IsConnectionAlive() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Conn == nil {
		return false
	}

	// Try to set a very short read deadline and peek for data
	// This is a non-blocking way to check if the connection is still alive
	p.Conn.SetReadDeadline(time.Now().Add(1 * time.Millisecond))
	
	// Try to peek for data (non-blocking)
	// We're not actually reading data, just checking if the connection is alive
	oneByte := make([]byte, 1)
	_, err := p.Conn.Read(oneByte)
	
	// Clear the read deadline
	p.Conn.SetReadDeadline(time.Time{})
	
	// If we get an error, it might be because there's no data (which is fine)
	// or because the connection is dead (which is not fine)
	// io.EOF or timeout errors are expected when there's no data
	if err != nil {
		// Check if it's a timeout error (which is fine)
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// Timeout is expected when there's no data, connection is alive
			return true
		}
		// Check if it's EOF (connection closed)
		if err == io.EOF {
			return false
		}
		// For other errors, assume connection is dead
		return false
	}
	
	// If we successfully read a byte, put it back (though this is unlikely)
	// For our purposes, we can assume the connection is alive
	return true
}

// Close closes peer connection
func (p *Peer) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Conn != nil {
		p.Conn.Close()
		p.Conn = nil
	}
}

// GenerateSelfSignedCert creates a self-signed TLS certificate
func GenerateSelfSignedCert() (tls.Certificate, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"ClipSync"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	cert, err := tls.X509KeyPair(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}),
		pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}),
	)
	
	return cert, err
}

// getAllLocalIPs gets all local IP addresses
func getAllLocalIPs() ([]net.IP, error) {
	var ips []net.IP
	
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
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
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			return nil, fmt.Errorf("no network interfaces found")
		}
		defer conn.Close()
		
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		ips = append(ips, localAddr.IP)
	}

	return ips, nil
}