//go:build !windows
// +build !windows

package clipboard

import (
	"context"
	"errors"
)

// Fallback manager for non-Windows systems
type FallbackManager struct {
	callback func(Content)
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewManager creates a new fallback clipboard manager
func NewManager() Manager {
	return &FallbackManager{}
}

func (f *FallbackManager) Start(ctx context.Context) error {
	f.ctx, f.cancel = context.WithCancel(ctx)
	f.running = true
	return nil
}

func (f *FallbackManager) Stop() error {
	if !f.running {
		return nil
	}

	if f.cancel != nil {
		f.cancel()
	}

	f.running = false
	return nil
}

func (f *FallbackManager) Get() (Content, error) {
	return Content{}, errors.New("clipboard access not implemented for this platform")
}

func (f *FallbackManager) Set(content Content) error {
	return errors.New("clipboard access not implemented for this platform")
}

func (f *FallbackManager) OnChange(callback func(Content)) {
	f.callback = callback
}
