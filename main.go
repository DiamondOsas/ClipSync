package main

import (
	"clipsync/internal"
	"clipsync/ping"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	internal.WG.Add(5)
	go internal.RegisterDevice(ctx)
	fmt.Println(1)
	go internal.BrowseForDevices(ctx)
	fmt.Println(2)
	go internal.Listen()
	fmt.Println(3)
	go func(ctx context.Context) {
		defer internal.WG.Done()
		for {
			devices := ping.Ping(internal.IP)
			if len(devices) > 0 {
				fmt.Println(internal.Devices{Ip: devices})
			}
			<-ctx.Done()
		}
	}(ctx)
	fmt.Println(4)
	for {
		if len(internal.IP) != 0 {
			go internal.ChangedClipbord(ctx)
		}
	}
	internal.WG.Wait()
}
