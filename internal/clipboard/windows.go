//go:build windows
// +build windows

package clipboard

import (
	"context"
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                         = windows.NewLazySystemDLL("user32.dll")
	kernel32                       = windows.NewLazySystemDLL("kernel32.dll")
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
	CF_UNICODETEXT = 13
	GMEM_MOVEABLE  = 0x0002
)

type WindowsManager struct {
	callback func(string)
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
}

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

func (w *WindowsManager) Get() (string, error) {
	if err := w.openClipboard(); err != nil {
		return "", err
	}
	defer w.closeClipboard()
	if available, _ := w.isFormatAvailable(CF_UNICODETEXT); available {
		return w.getText()
	}
	return "", fmt.Errorf("clipboard is empty or not text")
}

func (w *WindowsManager) Set(text string) error {
	for i := 0; i < 3; i++ {
		if err := w.setClipboardWithRetry(text); err == nil {
			return nil
		}
		if i < 2 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	return fmt.Errorf("failed to set clipboard after 3 attempts")
}

func (w *WindowsManager) setClipboardWithRetry(text string) error {
	if err := w.openClipboard(); err != nil {
		return err
	}
	defer w.closeClipboard()
	emptyClipboardProc.Call()
	return w.setText(text)
}

func (w *WindowsManager) OnChange(callback func(string)) {
	w.callback = callback
}

func (w *WindowsManager) monitorClipboard() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	var lastText string
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			if w.callback == nil {
				continue
			}
			text, err := w.Get()
			if err != nil || text == lastText {
				continue
			}
			lastText = text
			w.callback(text)
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

func (w *WindowsManager) getText() (string, error) {
	hMem, _, _ := getClipboardDataProc.Call(CF_UNICODETEXT)
	if hMem == 0 {
		return "", fmt.Errorf("GetClipboardData failed")
	}
	ptr, _, _ := globalLockProc.Call(hMem)
	if ptr == 0 {
		return "", fmt.Errorf("GlobalLock failed")
	}
	defer globalUnlockProc.Call(hMem)
	// Read UTF-16 string
	var buf []uint16
	for i := 0; ; i++ {
		ch := *(*uint16)(unsafe.Pointer(ptr + uintptr(i*2)))
		if ch == 0 {
			break
		}
		buf = append(buf, ch)
	}
	return windows.UTF16ToString(buf), nil
}

func (w *WindowsManager) setText(text string) error {
	utf16 := windows.StringToUTF16(text)
	size := len(utf16) * 2
	hMem, _, _ := globalAllocProc.Call(GMEM_MOVEABLE, uintptr(size))
	if hMem == 0 {
		return fmt.Errorf("GlobalAlloc failed")
	}
	ptr, _, _ := globalLockProc.Call(hMem)
	if ptr == 0 {
		globalFreeProc.Call(hMem)
		return fmt.Errorf("GlobalLock failed")
	}
	defer globalUnlockProc.Call(hMem)
	for i, v := range utf16 {
		*(*uint16)(unsafe.Pointer(ptr + uintptr(i*2))) = v
	}
	setClipboardDataProc.Call(CF_UNICODETEXT, hMem)
	return nil
}
