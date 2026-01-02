package ping

import (
	"fmt"
	"testing"
	"clipsync/modules"
)

func pingtest(t *testing.T){
	want := modules.Devices{Ip: []string{"ok"}}
	test := []string{"google.com", "192.168.23.4", "github.com"}
	output := ping(test)
	if output != want{
		t.Errorf("Imput: ",test, "Output", output,)
	}
	fmt.Println(output)
}
