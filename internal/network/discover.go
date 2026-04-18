package network

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"clipsync/internal/globals"
	"clipsync/internal/view"

	"github.com/grandcat/zeroconf"
)

var Entries = make(chan *zeroconf.ServiceEntry)

func getAllInterfaces() []net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var result []net.Interface
	for _, iface := range ifaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		// Skip interfaces with no addresses
		addrs, err := iface.Addrs()
		if err != nil || len(addrs) == 0 {
			continue
		}
		result = append(result, iface)
	}
	return result
}

func RegisterDevice(ctx context.Context, name string) error {
	if name == "" {
		globals.Username, _ = os.Hostname()
		name = globals.Username
	}

	ifaces := getAllInterfaces()

	server, err := zeroconf.Register(name, "_clipsync._tcp", "local.", globals.PORT, []string{""}, ifaces)

	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Broadcasting Presence...")
	defer server.Shutdown()
	<-ctx.Done()
	return nil
}

// Discover all services on the network (e.g. _workstation._tcp)

func BrowseForDevices(ctx context.Context) error {
	ifaces := getAllInterfaces()
	reslover, err := zeroconf.NewResolver(zeroconf.SelectIfaces(ifaces))

	if err != nil {
		log.Println(err)
		return err
	}

	go entry(Entries)

	err = reslover.Browse(ctx, "_clipsync._tcp", "local.", Entries)

	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Starting to Discover Services...")

	<-ctx.Done()
	return nil
}

func entry(results <-chan *zeroconf.ServiceEntry) {
	for entry := range results {
		if entry.Instance != globals.Username {
			newIP := string(entry.AddrIPv4[0].String())
			newDevice := globals.Device{Name: entry.HostName, Ip: newIP}
			go view.UpdateDevices(newDevice)
			globals.IPSMu.Lock()
			globals.IPS = append(globals.IPS, newIP)
			globals.IPSMu.Unlock()

			go Connect(newIP)
			log.Println("Found Device: Name: ", entry.Instance, " IP: ", entry.AddrIPv4)

			fmt.Println("Connected Device:", entry.Instance)
			
		 
		}
	}
}
