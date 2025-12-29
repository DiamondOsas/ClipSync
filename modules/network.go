package modules

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
)

var WG sync.WaitGroup
 
func RegisterDevice(ctx context.Context){
	defer WG.Done()
	//Add that when it display all the interfaces 
	log.Println("Starting to Register Device")
	instance, _ := os.Hostname()
	server, err :=zeroconf.Register(instance, "_clipsync._tcp","local.", 9999 , []string{""},nil)
	if err != nil{
		log.Fatal(err)
	}
	log.Println("Deivce Registered")
	defer server.Shutdown()

	<- ctx.Done()
}
// Discover all services on the network (e.g. _workstation._tcp)

func BrowseForDevices(ctx context.Context){
	defer WG.Done()
	log.Println("Starting to Discover Services")
	r, err := zeroconf.NewResolver(nil)
	if err != nil{
		log.Fatal(err)
	}
	entries := make( chan *zeroconf.ServiceEntry)
	go entry(ctx, entries)
	time , cancel := context.WithTimeout(context.Background(), time.Hour*100)

	
	defer cancel()

	err = r.Browse(time, "_clipsync._tcp", "local.", entries)
	if err != nil{
		log.Fatal(err)
	}
	<- ctx.Done()
}

func entry(ctx context.Context, results <-chan *zeroconf.ServiceEntry){
	for{
		select {
		case entry := <-results:
			log.Println("Found Device: ",entry.Instance, entry.AddrIPv4, entry.Instance, entry.Text)
		case <-ctx.Done():
			return 
		
		}
	}
}

