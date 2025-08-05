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
	port         int
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
		port:         port,
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

	// Start server message handler
	sm.wg.Add(1)
	go sm.serverMessageHandler()

	log.Printf("ClipSync started for device %s on port %d", sm.deviceID, sm.port)
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
func (sm *SyncManager) handleClipboardChange(text string) {
	if text == "" {
		return
	}
	
	msg := protocol.NewClipboardUpdate(sm.deviceID, text)
	sm.broadcast(msg)
}

// discoveryLoop handles device discovery
func (sm *SyncManager) discoveryLoop() {
	defer sm.wg.Done()

	entries, err := sm.discovery.Browse(sm.ctx)
	if err != nil {
		log.Printf("Discovery error: %v", err)
		return
	}

	for {
		select {
		case entry := <-entries:
			if entry == nil || sm.ctx.Err() != nil {
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

	ticker := time.NewTicker(30 * time.Second)
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

// serverMessageHandler handles messages from server clients
func (sm *SyncManager) serverMessageHandler() {
	defer sm.wg.Done()

	for {
		select {
		case msg := <-sm.server.Receive():
			if msg == nil || sm.ctx.Err() != nil {
				return
			}
			sm.handleServerMessage(msg)
		case <-sm.ctx.Done():
			return
		}
	}
}

// connectToDevice establishes connection to a discovered device
func (sm *SyncManager) connectToDevice(device *discovery.Device) {
	sm.clientLock.RLock()
	if _, exists := sm.clients[device.ID]; exists {
		sm.clientLock.RUnlock()
		return
	}
	sm.clientLock.RUnlock()

	client := network.NewClient(sm.deviceID, fmt.Sprintf("%s:%d", device.IP, device.Port))
	ctx, cancel := context.WithTimeout(sm.ctx, 5*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect to %s: %v", device.ID, err)
		return
	}

	// Send current clipboard
	if text, err := sm.clipboardMgr.Get(); err == nil && text != "" {
		client.SendMessage(protocol.NewClipboardUpdate(sm.deviceID, text))
	}

	sm.clientLock.Lock()
	sm.clients[device.ID] = client
	sm.clientLock.Unlock()

	// Start message handler for this client
	sm.wg.Add(1)
	go sm.handleClientMessages(device.ID, client)
}

// handleClientMessages processes messages from a client
func (sm *SyncManager) handleClientMessages(deviceID string, client *network.Client) {
	defer sm.wg.Done()

	for {
		select {
		case msg := <-client.Receive():
			if msg == nil || sm.ctx.Err() != nil {
				return
			}
			sm.handleMessage(deviceID, msg)
		case <-sm.ctx.Done():
			return
		}
	}
}

// handleServerMessage processes messages from server clients
func (sm *SyncManager) handleServerMessage(msg *protocol.Message) {
	if msg.Type == protocol.TypeClipboardUpdate && msg.DeviceID != sm.deviceID {
		sm.handleMessage(msg.DeviceID, msg)
	}
}

// handleMessage processes incoming clipboard updates
func (sm *SyncManager) handleMessage(deviceID string, msg *protocol.Message) {
	if msg.Type != protocol.TypeClipboardUpdate {
		return
	}

	// Prevent echo loops - skip if this is our own message
	if msg.DeviceID == sm.deviceID {
		return
	}

	// Update local clipboard
	if err := sm.clipboardMgr.Set(msg.Content); err != nil {
		log.Printf("Failed to set clipboard from %s: %v", deviceID, err)
	} else {
		log.Printf("Updated clipboard from %s: %s", deviceID, truncateText(msg.Content, 50))
	}
}

// broadcast sends message to all connected clients
func (sm *SyncManager) broadcast(msg *protocol.Message) {
	sm.clientLock.RLock()
	defer sm.clientLock.RUnlock()

	for deviceID, client := range sm.clients {
		if err := client.SendMessage(msg); err != nil {
			log.Printf("Failed to send to %s: %v", deviceID, err)
		}
	}
	sm.server.Broadcast(msg)
}

// cleanupClients removes disconnected clients
func (sm *SyncManager) cleanupClients() {
	sm.clientLock.Lock()
	defer sm.clientLock.Unlock()

	for deviceID, client := range sm.clients {
		if !client.IsConnected() {
			client.Disconnect()
			delete(sm.clients, deviceID)
			log.Printf("Disconnected from %s", deviceID)
		}
	}
}

// truncateText truncates text for logging
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
