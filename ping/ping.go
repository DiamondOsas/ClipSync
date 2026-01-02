package ping

import(
	"os/exec"
	"clipsync/modules"
	"sync"
)

func ping(ips []string) modules.Devices {
	var MU sync.RWMutex
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
			MU.Lock()
			activeips = append(activeips, val)
			MU.Unlock()
		modules.WG.Wait()
		}
		}(val)
	}

	return  modules.Devices{Ip: activeips}
}