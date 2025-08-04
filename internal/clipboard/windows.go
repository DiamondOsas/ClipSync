//go:build windows
// +build windows

package clipboard

import (
	"context"
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	// Removed unused addClipboardFormatListenerProc and removeClipboardFormatListenerProc
	openClipboardProc              = user32.NewProc("OpenClipboard")
	closeClipboardProc             = user32.NewProc("CloseClipboard")
	emptyClipboardProc             = user32.NewProc("EmptyClipboard")
	getClipboardDataProc           = user32.NewProc("GetClipboardData")
	setClipboardDataProc           = user32.NewProc("SetClipboardData")
	isClipboardFormatAvailableProc = user32.NewProc("IsClipboardFormatAvailable")
	globalAllocProc                = kernel32.NewProc("GlobalAlloc")
	globalFreeProc                 = kernel32.NewProc("GlobalFree")
	globalLockProc                 = kernel32.NewProc("GlobalLock")
	globalUnlockProc               = kernel32.NewProc("GlobalUnlock")
)

const (
	CF_TEXT            = 1
	CF_UNICODETEXT     = 13
	CF_HDROP           = 15
	CF_BITMAP          = 2
	CF_DIB             = 8
	WM_CLIPBOARDUPDATE = 0x031D
	GMEM_MOVEABLE      = 0x0002
	HWND_MESSAGE       = ^uintptr(2) + 1 // -3
)

// WindowsManager implements clipboard.Manager for Windows
type WindowsManager struct {
	callback func(Content)
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewManager creates a new Windows clipboard manager
func NewManager() Manager {
	return &WindowsManager{}
}

func (w *WindowsManager) Start(ctx context.Context) error {
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.running = true

	go w.monitorClipboard()
	return nil
}

func (w *WindowsManager) Stop() error {
	if !w.running {
		return nil
	}

	if w.cancel != nil {
		w.cancel()
	}

	w.running = false
	return nil
}

func (w *WindowsManager) Get() (Content, error) {
	if err := w.openClipboard(); err != nil {
		return Content{}, err
	}
	defer w.closeClipboard()

	// Check for text
	if available, _ := w.isFormatAvailable(CF_UNICODETEXT); available {
		return w.getText()
	}

	// Check for files (CF_HDROP)
	if available, _ := w.isFormatAvailable(CF_HDROP); available {
		return w.getFiles()
	}

	// Check for bitmap
	if available, _ := w.isFormatAvailable(CF_DIB); available {
		return w.getImage()
	}

	return Content{}, ErrEmptyClipboard
}

func (w *WindowsManager) Set(content Content) error {
	// Retry loop for flaky Windows clipboard API
	for i := 0; i < 3; i++ {
		if err := w.setClipboardWithRetry(content); err == nil {
			return nil
		}
		if i < 2 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	return fmt.Errorf("failed to set clipboard after 3 attempts")
}

func (w *WindowsManager) setClipboardWithRetry(content Content) error {
	if err := w.openClipboard(); err != nil {
		return err
	}
	defer w.closeClipboard()

	// Empty the clipboard first
	emptyClipboardProc.Call()

	switch content.Type {
	case TypeText:
		return w.setText(content.Data.(string))
	case TypeFiles:
		return w.setFiles(content.Data.([]string))
	default:
		return ErrUnsupportedFormat
	}
}

func (w *WindowsManager) OnChange(callback func(Content)) {
	w.callback = callback
}

func (w *WindowsManager) monitorClipboard() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var lastContent string

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			if w.callback == nil {
				continue
			}

			content, err := w.Get()
			if err != nil {
				continue
			}

			if content.Type == TypeText {
				current := content.Data.(string)
				if current != lastContent {
					lastContent = current
					w.callback(content)
				}
			}
		}
	}
}

func (w *WindowsManager) openClipboard() error {
	r1, _, err := openClipboardProc.Call(0)
	if r1 == 0 {
		return fmt.Errorf("OpenClipboard failed: %v", err)
	}
	return nil
}

func (w *WindowsManager) closeClipboard() error {
	r1, _, _ := closeClipboardProc.Call()
	if r1 == 0 {
		return fmt.Errorf("CloseClipboard failed")
	}
	return nil
}

func (w *WindowsManager) isFormatAvailable(format uint) (bool, error) {
	r1, _, _ := isClipboardFormatAvailableProc.Call(uintptr(format))
	return r1 != 0, nil
}

func (w *WindowsManager) getText() (Content, error) {
	hMem, _, _ := getClipboardDataProc.Call(CF_UNICODETEXT)
	if hMem == 0 {
		return Content{}, fmt.Errorf("GetClipboardData failed")
	}

	ptr, _, _ := globalLockProc.Call(hMem)
	if ptr == 0 {
		return Content{}, fmt.Errorf("GlobalLock failed")
	}
	defer globalUnlockProc.Call(hMem)

	// Limit the slice to the actual size to avoid unsafe.Pointer misuse
	utf16Slice := (*[1 << 16]uint16)(unsafe.Pointer(ptr))[:]
	text := syscall.UTF16ToString(utf16Slice)
	return Content{
		Type: TypeText,
		Data: text,
	}, nil
}

func (w *WindowsManager) setText(text string) error {
	utf16, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}

	size := uintptr(len(utf16) * 2)
	hMem, _, _ := globalAllocProc.Call(GMEM_MOVEABLE, size+2)
	if hMem == 0 {
		return fmt.Errorf("GlobalAlloc failed")
	}

	ptr, _, _ := globalLockProc.Call(hMem)
	if ptr == 0 {
		globalFreeProc.Call(hMem)
		return fmt.Errorf("GlobalLock failed")
	}

	// Limit the copy to the actual size to avoid unsafe.Pointer misuse
	dst := (*[1 << 16]byte)(unsafe.Pointer(ptr))[:size]
	src := (*[1 << 16]byte)(unsafe.Pointer(&utf16[0]))[:size]
	copy(dst, src)
	globalUnlockProc.Call(hMem)

	r1, _, _ := setClipboardDataProc.Call(CF_UNICODETEXT, hMem)
	if r1 == 0 {
		globalFreeProc.Call(hMem)
		return fmt.Errorf("SetClipboardData failed")
	}

	return nil
}

func (w *WindowsManager) getFiles() (Content, error) {
	// Simplified file handling - return empty for now
	return Content{}, ErrUnsupportedFormat
}

func (w *WindowsManager) getImage() (Content, error) {
	// Simplified image handling - return empty for now
	return Content{}, ErrUnsupportedFormat
}

func (w *WindowsManager) setFiles(_ []string) error {
	// Simplified file handling - return error for now
	return ErrUnsupportedFormat
}
