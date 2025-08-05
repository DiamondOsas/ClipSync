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

// Client represents a network client for clipboard sync
type Client struct {
	conn       net.Conn
	deviceID   string
	remoteAddr string
	sendChan   chan *protocol.Message
	recvChan   chan *protocol.Message
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	connected  bool
	mu         sync.RWMutex
}

// NewClient creates a new network client with default configuration
func NewClient(deviceID, remoteAddr string) *Client {
	return &Client{
		deviceID:   deviceID,
		remoteAddr: remoteAddr,
		sendChan:   make(chan *protocol.Message, 10),
		recvChan:   make(chan *protocol.Message, 10),
	}
}

// Connect establishes connection to remote device
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return fmt.Errorf("client already connected")
	}

	c.ctx, c.cancel = context.WithCancel(ctx)

	conn, err := net.DialTimeout("tcp", c.remoteAddr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.connected = true

	// Start send/receive goroutines
	c.wg.Add(2)
	go c.sendLoop()
	go c.recvLoop()

	return nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Disconnect closes the connection
func (c *Client) Disconnect() {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return
	}
	c.connected = false
	c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	c.wg.Wait()
}

// SendMessage sends a message to the remote device
func (c *Client) SendMessage(msg *protocol.Message) error {
	select {
	case c.sendChan <- msg:
		return nil
	case <-c.ctx.Done():
		return fmt.Errorf("client closed")
	}
}

// Receive returns the receive channel
func (c *Client) Receive() <-chan *protocol.Message {
	return c.recvChan
}

// sendLoop handles outgoing messages
func (c *Client) sendLoop() {
	defer c.wg.Done()
	for {
		select {
		case msg := <-c.sendChan:
			data, err := msg.Serialize()
			if err != nil {
				continue
			}
			data = append(data, '\n')
			c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			_, err = c.conn.Write(data)
			if err != nil {
				c.Disconnect()
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// recvLoop handles incoming messages
func (c *Client) recvLoop() {
	defer c.wg.Done()
	reader := bufio.NewReader(c.conn)
	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error: %v\n", err)
			}
			c.Disconnect()
			return
		}

		data = bytes.TrimSpace(data)
		if len(data) == 0 {
			continue
		}

		msg, err := protocol.DeserializeMessage(data)
		if err != nil {
			fmt.Printf("Failed to deserialize message: %v\n", err)
			continue
		}

		select {
		case c.recvChan <- msg:
		case <-c.ctx.Done():
			return
		}
	}
}
