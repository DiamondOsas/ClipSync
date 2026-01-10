package clipboard

import (
	"context"
	"log"
	"sync"
	"golang.design/x/clipboard"
)

func init() {
	err := clipboard.Init()
	if err != nil {
		log.Println(err)
	}
}
func CopyClipboard() string {
	data := clipboard.Read(clipboard.FmtText)
	return string(data)
}

func WriteClipboard(data string) {
	byte := []byte(data)
	clipboard.Write(clipboard.FmtText, byte)
}

//Find out how to check whether a clipboard fucntion forever below

func ChangedClipbord(ctx context.Context) bool {
	var mu sync.RWMutex
	defer WG.Done()
	changed := clipboard.Watch(context.TODO(), clipboard.FmtText)
	for info := range changed {
		str := string(info)
		if str == Recieved {
			continue
		} else {

			mu.Lock()
			data := CopyClipboard()
			same := (data == str)
			mu.Unlock()
			if same {
				//Test WriteClipboard("ok")
				SendClipboard()
				return true
			}
		}
	}
	<-ctx.Done()
	return false
}
