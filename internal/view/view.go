package view

import (
	"clipsync/gui"
	"clipsync/gui/pages"
	"clipsync/internal/globals"
)



// UpdateDevices handles adding a new device to both global and GUI state.
func UpdateDevices(Device globals.Device) {
	// 1. Update Global State
	globals.ConnDevicesMu.Lock()
	globals.ConnDevices = append(globals.ConnDevices, Device)
	globals.ConnDevicesMu.Unlock()

	// 2. Update GUI state if active
	if gui.State != nil {
		newDevice := pages.Device{
			Name: Device.Name,
			IP:   Device.Ip,
		}
		gui.State.Devices = append(gui.State.Devices, newDevice)
		RedrawUI()
	}
}

// UpdateClipboard handles adding new clipboard data to both global and GUI state.
func UpdateClipboard(data string) {
	if data == "" {
		return
	}

	// 1. Update Global State (Stack behavior: newest first)
	globals.ClipHistoryMu.Lock()
	globals.ClipHistory = append([]string{data}, globals.ClipHistory...)
	globals.ClipHistoryMu.Unlock()

	// 2. Update GUI state if active
	if gui.State != nil {
		gui.State.History = append([]string{data}, gui.State.History...)
		RedrawUI()
	}
}

func RedrawUI() {
	// Redraw the UI to show changes in both Update Devices and Clipboard
	if gui.Window != nil{
		gui.Window.Invalidate()
	}
}