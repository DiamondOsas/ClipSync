package modules

import (
	"log"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/atotto/clipboard"
)

var (
	clipboardMutex sync.Mutex
	lastContent    string
	lastReadTime   time.Time
	readCooldown   = 100 * time.Millisecond
)

// ReadClipboard reads from the system clipboard
func ReadClipboard() (string, error) {
	if clipboard.Unsupported {
		return "", nil
	}

	clipboardMutex.Lock()
	defer clipboardMutex.Unlock()

	if time.Since(lastReadTime) < readCooldown {
		return lastContent, nil
	}

	content, err := clipboard.ReadAll()
	if err != nil {
		log.Printf("Clipboard read error: %v", err)
		return "", err
	}

	if content == "" {
		return "", nil
	}

	converted := convertToSupportedFormat(content)
	if converted == "" {
		return "", nil
	}

	lastContent = converted
	lastReadTime = time.Now()
	return converted, nil
}

// WriteClipboard writes to the system clipboard
func WriteClipboard(text string) error {
	if clipboard.Unsupported {
		return nil
	}

	if text == "" {
		return nil
	}

	clipboardMutex.Lock()
	defer clipboardMutex.Unlock()

	if text == lastContent {
		return nil
	}

	converted := convertToSupportedFormat(text)
	if converted == "" {
		return nil
	}

	err := clipboard.WriteAll(converted)
	if err != nil {
		log.Printf("Clipboard write error: %v", err)
		return err
	}

	lastContent = converted
	lastReadTime = time.Now()
	return nil
}

// convertToSupportedFormat ensures text is valid UTF-8
func convertToSupportedFormat(content string) string {
	if !utf8.ValidString(content) {
		var buf strings.Builder
		for i := 0; i < len(content); {
			r, size := utf8.DecodeRuneInString(content[i:])
			if r == utf8.RuneError && size == 1 {
				buf.WriteRune('\uFFFD')
				i++
			} else {
				buf.WriteRune(r)
				i += size
			}
		}
		content = buf.String()
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	return content
}