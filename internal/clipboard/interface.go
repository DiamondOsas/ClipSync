package clipboard

import (
	"context"
	"errors"
)

var (
	ErrUnsupportedFormat = errors.New("unsupported clipboard format")
	ErrEmptyClipboard    = errors.New("clipboard is empty")
)

// Manager defines the interface for clipboard operations
type Manager interface {
	// Start begins monitoring the clipboard for changes
	Start(ctx context.Context) error
	
	// Stop stops monitoring the clipboard
	Stop() error
	
	// Get retrieves the current clipboard content
	Get() (Content, error)
	
	// Set sets the clipboard content
	Set(content Content) error
	
	// OnChange registers a callback for clipboard changes
	OnChange(callback func(Content))
}

// Content represents clipboard data
type Content struct {
	Type    ContentType
	Data    interface{}
	Formats []string
}

// ContentType defines the type of clipboard content
type ContentType string

const (
	TypeText  ContentType = "text"
	TypeImage ContentType = "image"
	TypeFiles ContentType = "files"
)

// IsValid checks if the content type is valid
func (ct ContentType) IsValid() bool {
	switch ct {
	case TypeText, TypeImage, TypeFiles:
		return true
	default:
		return false
	}
}
