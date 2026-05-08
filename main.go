package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"clipsync/gui"
	"clipsync/internal/cli"
	"clipsync/internal/clipboard"
	"clipsync/internal/core"
)

var Version = "dev"

func main() {
	clipboard.Init()

	// Intercept CLI execution. If it returns true, we shouldn't start GUI.
	if cli.Run() {
		return
	}

	// Setup context for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Run background sync tasks in a goroutine
	go func() {
		err := core.StartSync(ctx)
		if err != nil && err != context.Canceled {
			log.Printf("Background sync stopped: %v", err)
		}
	}()
	
	// Start the GUI (blocking call)
	gui.StartGUI()
}
