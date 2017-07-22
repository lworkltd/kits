package discovery

import (
	"fmt"
	"testing"
)

func TestInitDisconvery(t *testing.T) {
	InitDisconvery(&DiscoveryOption{
		ConsulHost: "10.25.100.164:8500",
	})
	key := "kits/unittest/hello"
	value, e := KeyValue("kits/unittest/hello")
	if e != nil || value != "world" {
		t.Errorf("key %s in consul,expect %v,get %s,err=%v", key, "world", value, e)
		return
	}

	o := &RegisterOption{
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
