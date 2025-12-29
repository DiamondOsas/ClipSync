package main

import (
	"clipsync/modules"
	"sync"
)
var wg sync.WaitGroup
func main() {
	wg.Add(1)
	go modules.RegisterDevice()
	go modules.BrowseForDevices()
 
	wg.Wait()
	// Block the main function from exiting, keeping the background services alive.
	
}
