package network

import (
	"bufio"
	"fmt"
	"log"

	// sysClipboard "golang.design/x/clipboard"
	"clipsync/internal/globals"
	appClipboard "clipsync/internal/clipboard"

)

func SendClipboard() {
	data := appClipboard.CopyClipboard()
	data = data + "\n"
	bytes := []byte(data)
	_, err := Conn.Write(bytes)
	if err != nil {
		log.Println(err)
	}
}

func RecieveClipboard() {
	for {
		conn, err := Ln.Accept()
		if err != nil {
			log.Println(err)
		}
		msg := bufio.NewReader(conn)
		globals.Recieved, _ = msg.ReadString('\n')
		if globals.Recieved == "Clipsync Here" {
			break
		} else {
			appClipboard.WriteClipboard(globals.Recieved)
			fmt.Println("Clipboard Updated")
		}
	}
}
