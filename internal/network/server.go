package network

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/DiamondOsas/ClipSync/internal/protocol"
)

// Server represents a network server for clipboard sync
type Server struct {
	listener   net.Listener
	deviceID   string
	port       int
	clients    map[string]*Client
	clientLock sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewServer creates a new network server
func NewServer(deviceID string, port int) *Server {
	return &Server{
		deviceID: deviceID,
		port:     port,
		clients:  make(map[string]*Client),
	}
}

// Start begins listening for connections
func (s *Server) Start(ctx context.Context) error {
	var err error
	s.ctx, s.cancel = context.WithCancel(ctx)

	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

// Stop shuts down the server
func (s *Server) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.listener != nil {
		s.listener.Close()
	}

	// Disconnect all clients
	s.clientLock.Lock()
	for _, client := range s.clients {
		client.Disconnect()
	}
	s.clients = make(map[string]*Client)
	s.clientLock.Unlock()

	s.wg.Wait()
}

// Broadcast sends a message to all connected clients
func (s *Server) Broadcast(msg *protocol.Message) {
	s.clientLock.RLock()
	defer s.clientLock.RUnlock()

	for _, client := range s.clients {
		if err := client.SendMessage(msg); err != nil {
			fmt.Printf("Failed to send to client: %v\n", err)
		}
	}
}

// GetClients returns a list of connected client IDs
func (s *Server) GetClients() []string {
	s.clientLock.RLock()
	defer s.clientLock.RUnlock()

	clients := make([]string, 0, len(s.clients))
	for id := range s.clients {
		clients = append(clients, id)
	}
	return clients
}

// acceptLoop handles incoming connections
func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.ctx.Err() == nil {
				fmt.Printf("Accept error: %v\n", err)
			}
			return
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection manages a single client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	// Read device hello message first
	msg, err := s.readMessage(conn)
	if err != nil {
		fmt.Printf("Failed to read hello: %v\n", err)
		return
	}

	if msg.Type != protocol.TypeDeviceHello {
		fmt.Printf("Expected device hello, got %s\n", msg.Type)
		return
	}

	clientID := msg.DeviceID
	client := NewClient(clientID, conn.RemoteAddr().String())
	client.conn = conn
	client.ctx, client.cancel = context.WithCancel(s.ctx)

	// Add client to map
	s.clientLock.Lock()
	s.clients[clientID] = client
	s.clientLock.Unlock()

	fmt.Printf("Client connected: %s\n", clientID)

	// Start client handlers
	client.wg.Add(2)
	go client.sendLoop()
	go client.recvLoop()

	// Wait for client to disconnect
	<-client.ctx.Done()

	// Remove client from map
	s.clientLock.Lock()
	delete(s.clients, clientID)
	s.clientLock.Unlock()

	fmt.Printf("Client disconnected: %s\n", clientID)
}

// readMessage reads a message from the connection
func (s *Server) readMessage(conn net.Conn) (*protocol.Message, error) {
	// Read length prefix
	var length uint32
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	// Read message data
	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	return protocol.DeserializeMessage(data)
}

// writeMessage writes a message to the connection
func (s *Server) writeMessage(conn net.Conn, msg *protocol.Message) error {
	data, err := msg.Serialize()
	if err != nil {
		return err
	}

	// Write length prefix
	length := uint32(len(data))
	if err := binary.Write(conn, binary.BigEndian, length); err != nil {
		return err
	}

	// Write message data
	_, err = conn.Write(data)
	return err
}
