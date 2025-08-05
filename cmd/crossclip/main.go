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

const (
	defaultPort = 8080
)

func main() {
	// Parse command line arguments
	deviceID := flag.String("device", "", "Unique device identifier (default: hostname)")
	port := flag.Int("port", defaultPort, "Port for communication")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Get device ID from hostname if not provided
	discovery.EnsureFirewall()

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to get hostname: %v", err)
	}

	if *deviceID == "" {
		*deviceID = hostname
	}

	// Enable debug logging if requested
	if *debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Create sync manager
	syncManager := sync.NewSyncManager(*deviceID, *port)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the sync manager
	if err := syncManager.Start(ctx); err != nil {
		log.Fatalf("Failed to start sync manager: %v", err)
	}

	fmt.Printf("ClipSync started on device: %s\n", *deviceID)
	fmt.Printf("Listening on port: %d\n", *port)
	fmt.Println("Press Ctrl+C to stop...")

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down...")

	// Graceful shutdown
	cancel()
	syncManager.Stop()

	fmt.Println("ClipSync stopped")
}
