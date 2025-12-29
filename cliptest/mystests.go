package main

import (
	"log"
	"net"
)

func main(){
	_, err := net.Interfaces()
	if err != nil{
		log.Fatal(err)
	}
	

	// for _, ifaces := range interfaces{
		
	// 		addr, _ := ifaces.Addrs()
	// 	for _, ip := range addr{
	// 	if ip.To
	// }
	// }


}