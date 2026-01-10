package ping

import (
	"clipsync/internal/network"
	"fmt"
	"reflect"
	"testing"
)

func TestPing(t *testing.T) {
	want := network.Devices{Ip: []string{"google.com", "github.com"}}
	test := []string{"google.com", "192.168.23.4", "github.com"}
	output := Ping(test)
	if !reflect.DeepEqual(output, want) {
		t.Errorf("got %v, want %v", output, want)
	}
	fmt.Println(output)
}
