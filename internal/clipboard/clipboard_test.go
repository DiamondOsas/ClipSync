package clipboard_test

import (
	"testing"
	"clipsync/internal/clipboard"
)

func TestClipboard(t *testing.T) {
	want := "Testing is taking place..."
	clipboard.WriteClipboard(want)
	output := clipboard.CopyClipboard()

	if want != output {
		t.Errorf("Input: %v Output : %v", want, output)
	}

}

// func TestChanged(t *testing.T){
// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	go ChangedClipbord()
// 	WriteClipboard("ok")

// }
