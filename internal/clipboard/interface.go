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
// Only supports text clipboard for Windows MVP
type Manager interface {
	// Start begins monitoring the clipboard for changes
	Start(ctx context.Context) error

	// Stop stops monitoring the clipboard
	Stop() error

	// Get retrieves the current clipboard content
	Get() (string, error)

	// Set sets the clipboard content
	Set(text string) error

	// OnChange registers a callback for clipboard changes
	OnChange(callback func(string))
}
