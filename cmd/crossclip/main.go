package main

import (
	"context"
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
	// Get device ID from hostname
	discovery.EnsureFirewall()

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to get hostname: %v", err)
	}

	// Create sync manager
	syncManager := sync.NewSyncManager(hostname, defaultPort)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the sync manager
	if err := syncManager.Start(ctx); err != nil {
		log.Fatalf("Failed to start sync manager: %v", err)
	}

	fmt.Printf("ClipSync started on device: %s\n", hostname)
	fmt.Printf("Listening on port: %d\n", defaultPort)
	fmt.Println("Press Ctrl+C to stop...")

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down...")

	// Graceful shutdown
	cancel()
	syncManager.Stop()

	fmt.Println("ClipSync stopped")
}
