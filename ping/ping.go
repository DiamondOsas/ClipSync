package ping

import(
	"os/exec"
	// "clipsync/modules"
	"sync"
)

func Ping(ips []string) []string{
	// defer modules.WG.Done()
	var MU sync.RWMutex
	var wg sync.WaitGroup
	var activeips []string
	if len(ips) == 0{
		return nil
	}
	for _, val := range ips{
		wg.Add(1)
		go func(ip string){
		defer wg.Done()
		cmd := exec.Command("ping", "-n", "1", "-l", "1", ip )
		err := cmd.Run()
		if err ==  nil{
			MU.Lock()
			activeips = append(activeips, val)
			MU.Unlock()
		}
		}(val)
	}
	wg.Wait()
	return activeips
}