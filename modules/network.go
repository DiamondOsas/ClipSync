package modules

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/grandcat/zeroconf"
)

var WG sync.WaitGroup
 
func RegisterDevice(){
	defer WG.Done()
	//Add that when it display all the interfaces 
	log.Println("Starting to Register Device")
	server, err :=zeroconf.Register("Clipsync", "_clipsync._tcp","local.", 9999 , []string{""},nil)
	if err != nil{
		log.Fatal(err)
	}
	log.Println("Deivce Registered")
	defer server.Shutdown()

	sig := make(chan os.Signal, 1)

	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
}
// Discover all services on the network (e.g. _workstation._tcp)

func BrowseForDevices(){
	defer WG.Done()
	log.Println("Starting to Discover Services")
	r, err := zeroconf.NewResolver(nil)
	if err != nil{
		log.Fatal(err)
	}
	entries := make( chan *zeroconf.ServiceEntry)
	go entry(entries)
	ctx , cancel := context.WithTimeout(context.Background(), time.Hour*100)

	
	defer cancel()

	err = r.Browse(ctx, "_clipsync._tcp", "local.", entries)
	if err != nil{
		log.Fatal(err)
	}
	<- ctx.Done()
}

func entry(results <-chan *zeroconf.ServiceEntry){
		for entry := range results {
			log.Println(entry)
		}
}

