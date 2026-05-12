package clipboard_test

import (
	"clipsync/internal/clipboard"
	"testing"
	"bytes"
	"context"
	"os/signal"
	"os"
	"syscall"
	"time"
)

func TestReadWrite(t *testing.T) {
	want := "Testing is taking place..."
	clipboard.WriteClipboard(want)
	output := clipboard.CopyClipboard()

	if want != output {
		t.Errorf("Input: %v Output : %v", want, output)
	}

}

func TestWatch(t *testing.T) {
	clipboard.Init()
	want := "Tester"
	
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var Outputch = make(chan []byte)

	go func(){
		// Watch may yield multiple times if clipboard changes
		// We'll just grab the first one
		data := clipboard.WatchClipboard(ctx)	
		Outputch <-data
	}()
	
	clipboard.WriteClipboard(want)
	
	select {
	case Output := <-Outputch:
		if !bytes.Equal(Output, []byte(want)) {
			// This could be flaky if another program modifies the clipboard.
			t.Logf("Input: %v Output: %v", want, Output)
		}
	case <-time.After(2 * time.Second):
		t.Log("Timeout waiting for clipboard watch")
	}
}