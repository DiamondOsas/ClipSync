package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"clipsync/Internal/clipboard"
	"clipsync/internal/globals"
	"clipsync/internal/network"
	"clipsync/internal/ping"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	globals.WG.Add(5)
	go network.RegisterDevice(ctx)
	fmt.Println(1)
	go network.BrowseForDevices(ctx)
	fmt.Println(2)
	go network.Listen()
	fmt.Println(3)
	go func(ctx context.Context) {
		defer globals.WG.Done()
		for {
			devices := ping.Ping(globals.IP)
			if len(devices) > 0 {
				fmt.Println(network.Devices{Ip: devices})
			}
			<-ctx.Done()
		}
	}(ctx)
	fmt.Println(4)
	for {
		if len(globals.IP) != 0 {
			go clipboard.ChangedClipbord(ctx)
		}
	}
	globals.WG.Wait()
}
