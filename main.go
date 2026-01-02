package main

import (
	"clipsync/modules"
	"clipsync/ping"
	"context"
	"os"
	"os/signal"
	"syscall"
	"fmt"
)

func main() {
	

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	modules.WG.Add(3)
	go modules.RegisterDevice(ctx)
	go modules.BrowseForDevices(ctx)
	go modules.Listen()
	go func(ctx context.Context){
		for {
		devices	:= ping.Ping(modules.IP)
		if len(devices) > 0{
			fmt.Println(modules.Devices{Ip: devices})	
		}
		<- ctx.Done()
		}
	}(ctx)
	modules.WG.Wait()


}
