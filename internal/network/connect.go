package network

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"

	"clipsync/internal/globals"
	"github.com/grandcat/zeroconf"
)

// type Info struct {

// 	ConnectedTo map[string]string
// 	Dialer 		bool
// }

var Conn net.Conn
var Ln net.Listener

func Connect(results *zeroconf.ServiceEntry) {

	entry := results
	log.Println("Connecting to", entry.Instance)
	Conn, err := net.Dial("tcp", string(entry.AddrIPv4[0].String()+":"+strconv.Itoa(globals.PORT)))
	if err != nil {
		log.Println(err)
	}

	//Send and recive confirm form server
	_, err = fmt.Fprintf(Conn, "Clipsync Here")
	if err != nil {
		log.Println(err)
	}
}

func Listen() {
	defer globals.WG.Done()
	port := ":" + strconv.Itoa(globals.PORT)
	Ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Listening...")
	for {
		conn, err := Ln.Accept()
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
