package ping

import(
	"os/exec"
	"clipsync/modules"
	"sync"
)

func ping(ips []string) modules.Devices {
	var mu sync.RWMutex
	var activeips []string
	if len(ips) == 0{
		return modules.Devices{}
	}
	for _, val := range ips{
		modules.WG.Add(1)
		go func(ip string){
		defer modules.WG.Done()

		cmd := exec.Command("ping", "-n", "1", "-l", "1", ip )
		err := cmd.Run()
		if err ==  nil{
			mu.Lock()
			activeips = append(activeips, val)
			mu.Unlock()
		modules.WG.Wait()
		}
		}(val)
	}

	return  modules.Devices{Ip: activeips}
}