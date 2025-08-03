package protocol

import (
	"encoding/json"
	"time"
)

// Message types for cross-clipboard sync
const (
	TypeClipboardUpdate = "clipboard_update"
	TypeDeviceHello     = "device_hello"
	TypeDeviceBye       = "device_bye"
	TypePing            = "ping"
	TypePong            = "pong"
)

// ClipboardData represents clipboard content
type ClipboardData struct {
	Type    string      `json:"type"`    // "text", "image", "files"
	Content interface{} `json:"content"` // string for text, []byte for image, []string for files
	Size    int         `json:"size"`
}

// Message is the envelope for all cross-clip communications
type Message struct {
	Type      string         `json:"type"`
	DeviceID  string         `json:"device_id"`
	Timestamp time.Time      `json:"timestamp"`
	Data      ClipboardData  `json:"data,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Serialize converts message to JSON bytes
func (m *Message) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// DeserializeMessage parses JSON bytes into Message
func DeserializeMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return &msg, err
}

// NewClipboardUpdate creates a new clipboard update message
func NewClipboardUpdate(deviceID string, data ClipboardData) *Message {
	return &Message{
		Type:      TypeClipboardUpdate,
		DeviceID:  deviceID,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
}

// NewDeviceHello creates a device announcement
func NewDeviceHello(deviceID string, metadata map[string]any) *Message {
	return &Message{
		Type:      TypeDeviceHello,
		DeviceID:  deviceID,
		Timestamp: time.Now().UTC(),
		Metadata:  metadata,
	}
}
