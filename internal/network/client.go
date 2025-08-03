package network

import (
	"context"
	"encoding/binary"
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
}

// NewClient creates a new network client
func NewClient(deviceID, remoteAddr string) *Client {
	return &Client{
		deviceID:   deviceID,
		remoteAddr: remoteAddr,
		sendChan:   make(chan *protocol.Message, 100),
		recvChan:   make(chan *protocol.Message, 100),
	}
}

// Connect establishes connection to remote device
func (c *Client) Connect(ctx context.Context) error {
	conn, err := net.DialTimeout("tcp", c.remoteAddr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.ctx, c.cancel = context.WithCancel(ctx)

	// Start send/receive goroutines
	c.wg.Add(2)
	go c.sendLoop()
	go c.recvLoop()

	return nil
}

// Disconnect closes the connection
func (c *Client) Disconnect() {
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
			if err := c.writeMessage(msg); err != nil {
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

	for {
		msg, err := c.readMessage()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error: %v\n", err)
			}
			return
		}

		select {
		case c.recvChan <- msg:
		case <-c.ctx.Done():
			return
		}
	}
}

// writeMessage writes a message to the connection
func (c *Client) writeMessage(msg *protocol.Message) error {
	data, err := msg.Serialize()
	if err != nil {
		return err
	}

	// Write length prefix
	length := uint32(len(data))
	if err := binary.Write(c.conn, binary.BigEndian, length); err != nil {
		return err
	}

	// Write message data
	_, err = c.conn.Write(data)
	return err
}

// readMessage reads a message from the connection
func (c *Client) readMessage() (*protocol.Message, error) {
	// Read length prefix
	var length uint32
	if err := binary.Read(c.conn, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	// Read message data
	data := make([]byte, length)
	if _, err := io.ReadFull(c.conn, data); err != nil {
		return nil, err
	}

	return protocol.DeserializeMessage(data)
}
