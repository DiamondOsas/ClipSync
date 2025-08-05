package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DiamondOsas/ClipSync/internal/discovery"
	"github.com/DiamondOsas/ClipSync/internal/sync"
)

func main() {
	var (
		deviceID = flag.String("device", "", "Unique device identifier")
		port     = flag.Int("port", 9901, "Port for communication")
	)
	flag.Parse()

	discovery.EnsureFirewall()

	if *deviceID == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatal("Failed to get hostname:", err)
		}
		*deviceID = hostname
	}

	fmt.Printf("Starting ClipSync - Device: %s, Port: %d\n", *deviceID, *port)

	syncManager := sync.NewSyncManager(*deviceID, *port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := syncManager.Start(ctx); err != nil {
		log.Fatal("Failed to start sync manager:", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	syncManager.Stop()
}
