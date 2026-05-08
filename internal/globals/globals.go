package globals

import (
	"sync"
)

var (
	IPSMu    sync.Mutex
	IPS      []string
	Recieved string
	PORT     = 9999
	Username string
)

type Device struct {
	Name string
	Ip   string
}

var (
	ConnDevicesMu sync.Mutex
	ConnDevices   []Device

	ClipHistoryMu sync.Mutex
	ClipHistory   []string
)