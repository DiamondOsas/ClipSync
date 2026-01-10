package globals

import (
	"sync"
)

var (
	WG       sync.WaitGroup
	Recieved string
	IP       []string
	PORT     = 9999
)
