package modules

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
)


var PORT = 9999
var Instance string
var Username string
var Entries = make( chan *zeroconf.ServiceEntry)
//Add that when it display all the interfaces 
//Make it to work on a perfect LAN Peer to Peer Setup
var WG sync.WaitGroup
 
func RegisterDevice(ctx context.Context){
	defer WG.Done()

	log.Println("Starting to Register Device")
	Username, _ = os.Hostname()
	server, err :=zeroconf.Register(Username, "_clipsync._tcp","local.", PORT , []string{""},nil)
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

	go entry(ctx, Entries)
	time , cancel := context.WithTimeout(context.Background(), time.Hour*100)

	
	defer cancel()

	err = r.Browse(time, "_clipsync._tcp", "local.",Entries)
	if err != nil{
		log.Fatal(err)
	}
	<- ctx.Done()
}

func entry(ctx context.Context, results <-chan *zeroconf.ServiceEntry){
	for{
		select {
		case entry := <-results:
			if entry.Instance == Username{
				continue
			}else{
				log.Println("Found Device: ",entry.Instance, entry.AddrIPv4, entry.Instance, entry.Text)
				Connect(results)
			}
			
		 
		
		//Connect function call
		
		case <-ctx.Done():
			return 	
		
		}
	}
}

