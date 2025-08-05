package protocol

import (
	"encoding/json"
	"time"
)

const (
	TypeClipboardUpdate = "clipboard_update"
	TypeDeviceHello     = "device_hello"
	TypeDeviceBye       = "device_bye"
	TypePing            = "ping"
	TypePong            = "pong"
	TypeAck             = "ack"
)

// Message is the envelope for all cross-clip communications
// For MVP, only text clipboard is supported
type Message struct {
	Type      string    `json:"type"`
	DeviceID  string    `json:"device_id"`
	Timestamp time.Time `json:"timestamp"`
	MessageID string    `json:"message_id,omitempty"`
	Content   string    `json:"content,omitempty"`
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
func NewClipboardUpdate(deviceID string, text string) *Message {
	return &Message{
		Type:      TypeClipboardUpdate,
		DeviceID:  deviceID,
		Timestamp: time.Now().UTC(),
		Content:   text,
	}
}

// NewDeviceHello creates a device announcement
func NewDeviceHello(deviceID string) *Message {
	return &Message{
		Type:      TypeDeviceHello,
		DeviceID:  deviceID,
		Timestamp: time.Now().UTC(),
	}
}

// NewAck creates an acknowledgment message
func NewAck(deviceID, messageID string) *Message {
	return &Message{
		Type:      TypeAck,
		DeviceID:  deviceID,
		Timestamp: time.Now().UTC(),
		MessageID: messageID,
	}
}
