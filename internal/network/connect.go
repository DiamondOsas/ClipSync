package network

import (
	// "bufio"
	"context"
	"encoding/binary"
	"log"
	"net"
	"strconv"

	"clipsync/internal/globals"
)

// type Info struct {

// 	ConnectedTo map[string]string
// 	Dialer 		bool
// }

var Conn *net.UDPConn
var Ready = make(chan struct{})

func Connect(ip string) {
	if Conn == nil {
		log.Println("Cannot connect, Conn is not initialized. Waiting for Ready channel...")
		<-Ready
	}
	addr, err := net.ResolveUDPAddr("udp", ip+":"+strconv.Itoa(globals.PORT))
	if err != nil {
		log.Println(err)
		return
	}
	msg := []byte("---ClipSync---")
	payload := make([]byte, 4+len(msg))
	binary.BigEndian.PutUint32(payload[:4], uint32(len(msg)))
	copy(payload[4:], msg)
	_, err = Conn.WriteToUDP(payload, addr)
	if err != nil {
		log.Println("Connect Write error:", err)
	}
}

func Listen(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(globals.PORT))
	if err != nil {
		log.Println(err)
		return err
	}
	Conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Listening For Connection...")
	close(Ready)

	<-ctx.Done()
	return nil
}


