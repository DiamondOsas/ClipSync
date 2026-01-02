package ping

import(
	"os/exec"
	"clipsync/modules"
)
func ping(ips []string) modules.Devices {
	if len(ips) == 0{
		return modules.Devices{}
	}
	for i, val := range ips{
		cmd := exec.Command("ping", "-n 1 -l 1", val )
		_, err := cmd.Output()
		if err !=  nil{
			newips := append(ips[:i], ips[i+1:]...)
			return modules.Devices{Ip: newips}
		}
	}

	return  modules.Devices{}
}