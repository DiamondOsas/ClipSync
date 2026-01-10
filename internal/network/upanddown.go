package internal

import (
	"bufio"
	"clipsync/internal/_clipboard"
	"fmt"
	"log"

	"golang.design/x/clipboard"
)

func SendClipboard() {
	data := internal._clipboard.CopyClipboard()
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
		Recieved, _ = msg.ReadString('\n')
		if Recieved == "Clipsync Here" {
			break
		} else {
			WriteClipboard(Recieved)
			fmt.Println("Clipboard Updated")
		}
	}
}
