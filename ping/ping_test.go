package ping

import (
	"fmt"
	"testing"
	"clipsync/modules"
	"reflect"
)

func TestPing(t *testing.T){
	want := modules.Devices{Ip: []string{"google.com","github.com"}}
	test := []string{"google.com", "192.168.23.4", "github.com"}
	output := ping(test)
	if !reflect.DeepEqual(output, want) {
    t.Errorf("got %v, want %v", output, want)
	}
	fmt.Println(output)
}
