package modules

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/grandcat/zeroconf"
)
type State struct{
	ConnectedTo []string
}

func Connect(results <- chan*zeroconf.ServiceEntry){
	 for entry := range results{
		log.Println("Trying to Connect to", entry.Instance)
		conn, err :=net.Dial("tcp", string(entry.AddrIPv4[0].String() + ":" + strconv.Itoa(PORT)))
		if err != nil{
			log.Fatal(err)
		}	
		

		//Send and recive confirm form server
		fmt.Fprintf(conn, "Clipsync Here")
		fmt.Println("Sending Handshake ")
		reply := bufio.NewReader(conn)
		data, err := reply.ReadString('\n')
		if err != nil{
			log.Fatal(err)
		}
		if data == "I Hear U"{
			fmt.Println("Recived HandShake")
		}

	}

}

func Listen(){
	defer WG.Done()
	port := ":" +  strconv.Itoa(PORT)
	ln, err := net.Listen("tcp", port)
	if err != nil{
		log.Fatal(err)
	}
	fmt.Println("Listening...")
	for{
		conn, err := ln.Accept(); if err != nil{log.Fatal(err)}
		fmt.Println("Recived Connection")
		msg := bufio.NewReader(conn)
		message,_ := msg.ReadString('\n')
		if message == "Clipsync Here"{
			conn.Write([]byte("I Hear U"))
			fmt.Println("Responded to ")
		}
	
	
	}

}