package network

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/DiamondOsas/ClipSync/internal/protocol"
)

// Server represents a network server for clipboard sync
type Server struct {
	listener   net.Listener
	deviceID   string
	port       int
	clients    map[string]*clientConnection
	clientLock sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	receiveCh  chan *protocol.Message
}

type clientConnection struct {
	conn     net.Conn
	deviceID string
	sendCh   chan *protocol.Message
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewServer creates a new network server
func NewServer(deviceID string, port int) *Server {
	return &Server{
		deviceID:  deviceID,
		port:      port,
		clients:   make(map[string]*clientConnection),
		receiveCh: make(chan *protocol.Message, 100),
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
		client.cancel()
		if client.conn != nil {
			client.conn.Close()
		}
	}
	s.clients = make(map[string]*clientConnection)
	s.clientLock.Unlock()

	s.wg.Wait()
	close(s.receiveCh)
}

// Broadcast sends a message to all connected clients
func (s *Server) Broadcast(msg *protocol.Message) {
	s.clientLock.RLock()
	defer s.clientLock.RUnlock()

	for _, client := range s.clients {
		select {
		case client.sendCh <- msg:
		default:
			// Client is slow, skip
		}
	}
}

// Receive returns the channel for incoming messages
func (s *Server) Receive() <-chan *protocol.Message {
	return s.receiveCh
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

	clientID := conn.RemoteAddr().String()
	ctx, cancel := context.WithCancel(s.ctx)
	
	client := &clientConnection{
		conn:     conn,
		deviceID: clientID,
		sendCh:   make(chan *protocol.Message, 10),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Add client to map
	s.clientLock.Lock()
	s.clients[clientID] = client
	s.clientLock.Unlock()

	// Start send/receive goroutines
	go s.clientSendLoop(client)
	s.clientReceiveLoop(client)

	// Cleanup
	s.clientLock.Lock()
	delete(s.clients, clientID)
	s.clientLock.Unlock()
	cancel()
	close(client.sendCh)
}

// clientSendLoop handles sending messages to a client
func (s *Server) clientSendLoop(client *clientConnection) {
	for {
		select {
		case msg := <-client.sendCh:
			data, err := msg.Serialize()
			if err != nil {
				continue
			}
			data = append(data, '\n')
			client.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			_, err = client.conn.Write(data)
			if err != nil {
				return
			}
		case <-client.ctx.Done():
			return
		}
	}
}

// clientReceiveLoop handles receiving messages from a client
func (s *Server) clientReceiveLoop(client *clientConnection) {
	reader := bufio.NewReader(client.conn)
	for {
		select {
		case <-client.ctx.Done():
			return
		default:
			client.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
			data, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Read error from client %s: %v\n", client.deviceID, err)
				}
				return
			}

			data = bytes.TrimSpace(data)
			if len(data) == 0 {
				continue
			}

			msg, err := protocol.DeserializeMessage(data)
			if err != nil {
				fmt.Printf("Failed to deserialize message from %s: %v\n", client.deviceID, err)
				continue
			}

			// Send to receive channel
			select {
			case s.receiveCh <- msg:
			case <-client.ctx.Done():
				return
			default:
				// Channel full, skip message
			}
		}
	}
}
