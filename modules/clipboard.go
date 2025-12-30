package modules

import (
	"context"
	"log"

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

func WriteClipboard(data string) bool {
	byte := []byte(data)
	write := clipboard.Write(clipboard.FmtText, byte)
	select {
	case <-write:
		return false
	}

}

func ChangedClipbord(){
	changed := clipboard.Watch(context.TODO(), clipboard.FmtText)
	for info := range changed {
		str := string(info)
		data := CopyClipboard()
		if data == str {
			SendClipboard()
		}
	}

}
