package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
)

var IP []string
var PORT = 9999
var Instance string
var Username string
var Entries = make(chan *zeroconf.ServiceEntry)
var Recieved string

// Add that when it display all the interfaces
// Make it to work on a perfect LAN Peer to Peer Setup
var WG sync.WaitGroup

func RegisterDevice(ctx context.Context) {
	defer WG.Done()

	log.Println("Starting to Register Device")
	Username, _ = os.Hostname()
	server, err := zeroconf.Register(Username, "_clipsync._tcp", "local.", PORT, []string{""}, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Deivce Registered")
	defer server.Shutdown()

	<-ctx.Done()
}

// Discover all services on the network (e.g. _workstation._tcp)

func BrowseForDevices(ctx context.Context) {
	defer WG.Done()
	log.Println("Starting to Discover Services")
	r, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Println(err)
	}

	go entry(ctx, Entries)
	time, cancel := context.WithTimeout(context.Background(), time.Hour*100)

	defer cancel()

	err = r.Browse(time, "_clipsync._tcp", "local.", Entries)
	if err != nil {
		log.Println(err)
	}
	<-ctx.Done()
}

func entry(ctx context.Context, results <-chan *zeroconf.ServiceEntry) {
	for {
		select {
		case entry := <-results:
			if entry.Instance == Username {
				continue
			} else {
				log.Println("Found Device: ", entry.Instance, entry.AddrIPv4, entry.Text)
				IP = append(IP, string(entry.AddrIPv4[0].String()))
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
