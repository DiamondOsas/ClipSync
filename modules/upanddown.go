package modules

import (
	"fmt"
	"log"
	"bufio"
)


func SendClipboard(){
	data := CopyClipboard()
	data = data + "\n"
	bytes := []byte(data)
	_, err :=Conn.Write(bytes)
	if err != nil{
		log.Println(err)
	}
}

func RecieveClipboard(){
	for {
		conn, err := Ln.Accept()
		if err != nil {
			log.Println(err)
		}
		msg := bufio.NewReader(conn)
		message, _ := msg.ReadString('\n')
		if message == "Clipsync Here"{
			break
		}else{
			WriteClipboard(message)
			fmt.Println("Clipboard Updated")
		}
	}
}
