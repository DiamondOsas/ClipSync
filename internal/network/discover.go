package network

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"clipsync/internal/globals"
	"github.com/grandcat/zeroconf"
)

var Instance string
var Entries = make(chan *zeroconf.ServiceEntry)

// Add that when it display all the interfaces
// Make it to work on a perfect LAN Peer to Peer Setup

func RegisterDevice(ctx context.Context) {
	defer globals.WG.Done()
	log.Println("Registering Device")
	globals.Username, _ = os.Hostname()


	server, err := zeroconf.Register(globals.Username, "_clipsync._tcp", "local.", globals.PORT, []string{""}, nil)
	
	if err != nil {
		log.Println("Could not Register Device Please Make sure you are connected to a net")
		log.Println(err)

	}
	
	log.Println("Deivce Registered")
	defer server.Shutdown()
	
	<-ctx.Done()
}

// Discover all services on the network (e.g. _workstation._tcp)

func BrowseForDevices(ctx context.Context) {
	defer globals.WG.Done()
	log.Println("Starting to Discover Services")
	reslover, err := zeroconf.NewResolver(nil)
	
	if err != nil {
		log.Println(err)
	}

	go entry(ctx, Entries)
	time, cancel := context.WithTimeout(context.Background(), time.Hour*100)

	defer cancel()

	err = reslover.Browse(time, "_clipsync._tcp", "local.", Entries)
	
	if err != nil {
		log.Println(err)
	}
	
	<-ctx.Done()
}

func entry(ctx context.Context, results <-chan *zeroconf.ServiceEntry) {
	for {
		select {
		case entry := <-results:
			if entry.Instance == globals.Username {
				continue
			} else {
				log.Println("Found Device: ", entry.Instance, entry.AddrIPv4, entry.Text)
				globals.IP = append(globals.IP, string(entry.AddrIPv4[0].String()))
				fmt.Println("Connected Device:", entry.Instance)

				// Connect(entry)
				// Discovered(entry.Instance ,entry.AddrIPv4[0].String())
			}

		//Connect function call

		case <-ctx.Done():
			return

		}
	}
}

func Discovered(name string, ip string) {
	// var arrnames map[string]string
	// arrnames[name] = ip

	// Info{arrnames, true}
}
