package main

import (
	"clipsync/modules"
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	modules.WG.Add(2)
	go modules.RegisterDevice(ctx)
	go modules.BrowseForDevices(ctx)
	go modules.Listen()
	modules.WG.Wait()
}
