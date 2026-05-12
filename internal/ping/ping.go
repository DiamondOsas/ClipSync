package ping

import (
	"log"
	"time"

	"github.com/go-ping/ping"
)

// PingIPS takes a list of IPs and returns only those that are reachable (at least 1 packet received).
func PingIPS(ips []string) []string {
	if ips == nil {
		time.Sleep(2 * time.Second)
		return nil
	}

	var reachable []string
	for _, ip := range ips {
		pinger, err := ping.NewPinger(ip)
		if err != nil {
			log.Println(err)
		}

		pinger.Count = 2 //Just incase the first one drops 
		pinger.Timeout = 2 * time.Second
		
		// SetPrivileged allows pinger to work on most systems without root
		pinger.SetPrivileged(true) 

		err = pinger.Run()
		if err != nil {
			log.Print(err)
		}

		stats := pinger.Statistics()
		if stats.PacketsRecv > 0 {
			reachable = append(reachable, ip)
		} 
	}
	return reachable
}