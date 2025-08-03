package sync

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/DiamondOsas/ClipSync/internal/clipboard"
	"github.com/DiamondOsas/ClipSync/internal/discovery"
	"github.com/DiamondOsas/ClipSync/internal/network"
	"github.com/DiamondOsas/ClipSync/internal/protocol"
)

// SyncManager coordinates clipboard synchronization between devices
type SyncManager struct {
	deviceID     string
	clipboardMgr clipboard.Manager
	discovery    *discovery.Discovery
	server       *network.Server
	clients      map[string]*network.Client
	clientLock   sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewSyncManager creates a new sync manager
func NewSyncManager(deviceID string, port int) *SyncManager {
	return &SyncManager{
		deviceID:     deviceID,
		clipboardMgr: clipboard.NewManager(),
		discovery:    discovery.New(deviceID, port),
		server:       network.NewServer(deviceID, port),
		clients:      make(map[string]*network.Client),
	}
}

// Start begins clipboard synchronization
func (sm *SyncManager) Start(ctx context.Context) error {
	sm.ctx, sm.cancel = context.WithCancel(ctx)

	// Start clipboard manager
	if err := sm.clipboardMgr.Start(sm.ctx); err != nil {
		return fmt.Errorf("failed to start clipboard manager: %w", err)
	}

	// Start discovery service
	if err := sm.discovery.Register(sm.ctx); err != nil {
		return fmt.Errorf("failed to start discovery: %w", err)
	}

	// Start network server
	if err := sm.server.Start(sm.ctx); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Set up clipboard change handler
	sm.clipboardMgr.OnChange(sm.handleClipboardChange)

	// Start discovery listener
	sm.wg.Add(1)
	go sm.discoveryLoop()

	// Start client connection manager
	sm.wg.Add(1)
	go sm.clientManagerLoop()

	log.Printf("ClipSync started for device %s", sm.deviceID)
	return nil
}

// Stop shuts down the sync manager
func (sm *SyncManager) Stop() {
	if sm.cancel != nil {
		sm.cancel()
	}

	sm.clipboardMgr.Stop()
	sm.discovery.Unregister()
	sm.server.Stop()

	// Disconnect all clients
	sm.clientLock.Lock()
	for _, client := range sm.clients {
		client.Disconnect()
	}
	sm.clients = make(map[string]*network.Client)
	sm.clientLock.Unlock()

	sm.wg.Wait()
	log.Println("ClipSync stopped")
}

// handleClipboardChange is called when local clipboard changes
func (sm *SyncManager) handleClipboardChange(content clipboard.Content) {
	dataStr, ok := content.Data.(string)
	if !ok {
		log.Printf("Clipboard content is not text, skipping sync")
		return
	}

	msg := &protocol.Message{
		Type:      protocol.TypeClipboardUpdate,
		DeviceID:  sm.deviceID,
		Timestamp: time.Now(),
		Data: protocol.ClipboardData{
			Type:    string(content.Type),
			Content: dataStr,
			Size:    len(dataStr),
		},
	}

	// Broadcast to all connected clients
	sm.server.Broadcast(msg)
}

// discoveryLoop handles device discovery
func (sm *SyncManager) discoveryLoop() {
	defer sm.wg.Done()

	entries, err := sm.discovery.Browse(sm.ctx)
	if err != nil {
		log.Printf("Failed to start discovery browsing: %v", err)
		return
	}

	for {
		select {
		case entry := <-entries:
			if entry == nil {
				return
			}
			device := discovery.ParseServiceEntry(entry)
			if device != nil && device.ID != sm.deviceID {
				go sm.connectToDevice(device)
			}
		case <-sm.ctx.Done():
			return
		}
	}
}

// clientManagerLoop manages client connections
func (sm *SyncManager) clientManagerLoop() {
	defer sm.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.cleanupClients()
		case <-sm.ctx.Done():
			return
		}
	}
}

// connectToDevice establishes connection to a discovered device
func (sm *SyncManager) connectToDevice(device *discovery.Device) {
	sm.clientLock.Lock()
	defer sm.clientLock.Unlock()

	// Skip if already connected
	if _, exists := sm.clients[device.ID]; exists {
		return
	}

	client := network.NewClient(sm.deviceID, fmt.Sprintf("%s:%d", device.IP, device.Port))
	ctx, cancel := context.WithTimeout(sm.ctx, 5*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect to %s: %v", device.ID, err)
		return
	}

	// Send device hello
	hello := &protocol.Message{
		Type:      protocol.TypeDeviceHello,
		DeviceID:  sm.deviceID,
		Timestamp: time.Now(),
	}
	if err := client.SendMessage(hello); err != nil {
		log.Printf("Failed to send hello to %s: %v", device.ID, err)
		client.Disconnect()
		return
	}

	// Start message handler
	sm.clients[device.ID] = client
	go sm.handleClientMessages(device.ID, client)
}

// handleClientMessages processes messages from a client
func (sm *SyncManager) handleClientMessages(deviceID string, client *network.Client) {
	for {
		select {
		case msg := <-client.Receive():
			if msg == nil {
				return
			}
			sm.handleMessage(deviceID, msg)
		case <-sm.ctx.Done():
			return
		}
	}
}

// handleMessage processes incoming messages
func (sm *SyncManager) handleMessage(deviceID string, msg *protocol.Message) {
	switch msg.Type {
	case protocol.TypeClipboardUpdate:
		if msg.Data.Type == "" {
			log.Printf("Invalid clipboard data from %s", deviceID)
			return
		}

		// Update local clipboard
		content := clipboard.Content{
			Type: clipboard.ContentType(msg.Data.Type),
			Data: msg.Data.Content.(string),
		}
		if err := sm.clipboardMgr.Set(content); err != nil {
			log.Printf("Failed to set clipboard: %v", err)
		}

	case protocol.TypeDeviceHello:
		log.Printf("Received hello from %s", deviceID)

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// cleanupClients removes disconnected clients
func (sm *SyncManager) cleanupClients() {
	sm.clientLock.Lock()
	defer sm.clientLock.Unlock()

	for deviceID, client := range sm.clients {
		// Simple check - if the client is nil or connection is closed
		// In a real implementation, you'd want more sophisticated health checks
		if client == nil {
			delete(sm.clients, deviceID)
			continue
		}
	}
}

// GetConnectedDevices returns a list of connected device IDs
func (sm *SyncManager) GetConnectedDevices() []string {
	sm.clientLock.RLock()
	defer sm.clientLock.RUnlock()

	devices := make([]string, 0, len(sm.clients))
	for id := range sm.clients {
		devices = append(devices, id)
	}
	return devices
}
