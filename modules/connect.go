package modules

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/grandcat/zeroconf"
)

type Info struct {
	ConnectedTo map[string]string
	Dialer 		bool
}

var Conn net.Conn

func Connect(results *zeroconf.ServiceEntry) {
	
		entry := results
		log.Println("Connecting to", entry.Instance)
		Conn, err := net.Dial("tcp", string(entry.AddrIPv4[0].String()+":"+strconv.Itoa(PORT)))
		if err != nil {
			log.Println(err)
		}

		//Send and recive confirm form server
		_, err = fmt.Fprintf(Conn, "Clipsync Here")
		if err != nil{
			log.Println(err)
		}




}

func Listen() {
	defer WG.Done()
	port := ":" + strconv.Itoa(PORT)
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Listening...")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}
		fmt.Println("Recived Connection")
		msg := bufio.NewReader(conn)
		message, _ := msg.ReadString('\n')
		if message == "Clipsync Here" {
			conn.Write([]byte("I Hear U"))
			fmt.Println("Responded")
		}

	}

}

func ping(){
	
}

func SendClipboard(){
	data := CopyClipboard()
	Conn.Write("b []byte")
}

func RecieveClipboard(){


}


