package discovery

import (
	"fmt"
	"testing"

	"github.com/lvhuat/kits/helper/consul"
)

func TestInitDisconvery(t *testing.T) {
	csl, err := consul.New("10.25.100.164:8500")
	Init(&Option{
		ConsulClient: csl,
	})
	key := "kits/unittest/hello"
	value, _, e := csl.KeyValue("kits/unittest/hello")
	if e != nil || value != "world" {
		t.Errorf("key %s in consul,expect %v,get %s,err=%v", key, "world", value, e)
		return
	}

	o := &consul.RegisterOption{
		Name: "kits-test-server",
		Id:   "kits-test-server-001",
		Ip:   "localhost",
		Port: 11111,
	}

	Register(o)
	remotes, err := Discover(o.Name)
	if err != nil || len(remotes) != 1 {
		t.Errorf("expect 1 server got %v ,err=%v", len(remotes), err)
	}
	if remotes[0] != fmt.Sprintf("%s:%d", o.Ip, o.Port) {
		t.Errorf("expect localhost:11111 server got %v", remotes[0])
	}

	Unregister(o)

	remotes, err = Discover(o.Name)
	if err != nil || len(remotes) != 1 {
		t.Errorf("expect 0 server got %v ,err=%v", len(remotes), err)
	}
}
