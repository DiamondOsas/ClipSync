package discovery

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/grandcat/zeroconf"
)

const (
	ServiceName = "_crossclip._tcp"
	Domain      = "local."
)

// Discovery handles mDNS service discovery
type Discovery struct {
	server   *zeroconf.Server
	resolver *zeroconf.Resolver
	deviceID string
	port     int
}

// New creates a new discovery instance
func New(deviceID string, port int) *Discovery {
	return &Discovery{
		deviceID: deviceID,
		port:     port,
	}
}

// Register announces this device on the network
func (d *Discovery) Register(ctx context.Context) error {
	service, err := zeroconf.Register(
		d.deviceID,        // service instance name
		ServiceName,       // service type
		Domain,            // domain
		d.port,            // port
		[]string{"Cross-Clip clipboard sync"}, // metadata
		nil,               // interfaces
	)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	d.server = service
	return nil
}

// Unregister stops announcing this device
func (d *Discovery) Unregister() {
	if d.server != nil {
		d.server.Shutdown()
	}
}

// Browse finds other Cross-Clip devices on the network
func (d *Discovery) Browse(ctx context.Context) (<-chan *zeroconf.ServiceEntry, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create resolver: %w", err)
	}

	d.resolver = resolver

	entries := make(chan *zeroconf.ServiceEntry)
	go func() {
		err := resolver.Browse(ctx, ServiceName, Domain, entries)
		if err != nil {
			log.Printf("Browse error: %v", err)
			close(entries)
		}
	}()

	return entries, nil
}

// GetLocalIP returns the local IP address for this device
func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// Device represents a discovered Cross-Clip device
type Device struct {
	ID       string
	Hostname string
	IP       string
	Port     int
	LastSeen time.Time
}

// ParseServiceEntry converts zeroconf entry to Device
func ParseServiceEntry(entry *zeroconf.ServiceEntry) *Device {
	if len(entry.AddrIPv4) == 0 {
		return nil
	}

	return &Device{
		ID:       entry.Instance,
		Hostname: entry.HostName,
		IP:       entry.AddrIPv4[0].String(),
		Port:     entry.Port,
		LastSeen: time.Now(),
	}
}
