package main

import (
	"clipsync/modules"

)

func main() {
	modules.WG.Add(2)
	go modules.RegisterDevice()
	go modules.BrowseForDevices()
	modules.WG.Wait()

	// Block the main function from exiting, keeping the background services alive.
	
}
