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

func WriteClipboard(data string) {
	byte := []byte(data)
	clipboard.Write(clipboard.FmtText, byte)

}

//Find out how to check whether a clipboard fucntion forever below

func ChangedClipbord()bool {
	defer WG.Done()
	changed := clipboard.Watch(context.TODO(), clipboard.FmtText)
	for info := range changed {
		str := string(info)
		if str == Recieved {
			break
		} else {
			data := CopyClipboard()
			if data == str {
				WriteClipboard("ok") //Test
				// SendClipboard()
			}
		}
	}
	return true
}
