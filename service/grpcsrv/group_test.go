package grpcsrv

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/lworkltd/kits/service/grpcinvoke"

	_ "github.com/lworkltd/kits/service/grpcinvoke/invokeimpl"
	"github.com/lworkltd/kits/service/grpcsrv/example/testproto"
	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
)

func TestInterfaceGroupSubGroup(t *testing.T) {
	var sum string
	service := newService()
	addabc := service.Group("add-abc", func(context.Context, *grpccomm.CommRequest) error {
		sum += "abc"
		return nil
	})
	group := service.groups["add-abc"]
	if group == nil {
		t.Errorf("group add-abc has register on service,but not found")
		return
	}
	if group.Name() != "add-abc" {
		t.Errorf("group add-abc has register on service,but group.Name() not return as expect,got %v", group.Name())
		return
	}

	add123 := addabc.Group("add-123", func(context.Context, *grpccomm.CommRequest) error {
		sum += "123"
		return nil
	})

	add123x := service.Group("add-123")
	if !reflect.DeepEqual(add123, add123x) {
		t.Errorf("unexpect equal group=%v,got %v", add123, add123x)
		return
	}

	addxyz := add123.Group("add-xyz", func(context.Context, *grpccomm.CommRequest) error {
		sum += "xyz"
		return nil
	})

	if err := addxyz.doPipe(nil, nil); err != nil {
		t.Errorf("unexpect error when do doPipe,%v", err)
		return
	}

	if sum != "abc123xyz" {
		t.Errorf("doPipe expect sum=abc123xyz,go %v", sum)
		return
	}

	resultErr := add123.Group("add-123", func(context.Context, *grpccomm.CommRequest) error {
		return fmt.Errorf("error")
	})
	err := resultErr.doPipe(nil, nil)
	if err == nil {
		t.Errorf("expect result error")
		return
	}
}

type TestUnsupportTarget struct {
}

func TestInterfaceGroupProxyRegister(t *testing.T) {
	type args struct {
		detect ProxyDetect
	}
	service := newService()
	group := service.Group("proxy-group")
	group.Proxy(func(context.Context, *grpccomm.CommRequest) (bool, grpcinvoke.Service, error) {
		return true, grpcinvoke.Name("abc123"), nil
	})

	matched, srv, err := service.proxyRules[0].target.detect(nil, nil)
	if !matched {
		t.Errorf("group.Proxy matched")
		return
	}
	if err != nil {
		t.Errorf("group.Proxy expect no error,got %v", err)
		return
	}
	if srv.Name() != "abc123" {
		t.Errorf("group.Proxy unexpect service name,got %v", srv.Name())
		return
	}

	var dTarget ProxyDetect = func(context.Context, *grpccomm.CommRequest) (bool, grpcinvoke.Service, error) {
		return false, nil, nil
	}
	group.Proxy(dTarget)
	group.ProxyInterface(grpcinvoke.Addr("127.0.0.8090"), &testproto.DepositRequest{})
	group.ProxyService("Calculator", grpcinvoke.Addr("127.0.0.8090"), &testproto.AddRequest{})
	group.ProxyService("Calculator", grpcinvoke.Addr("127.0.0.8090"), "EchoRequest")
	group.ProxyService("Bridge", grpcinvoke.Addr("127.0.0.8091"))
	group.ProxyService("Kzz", grpcinvoke.Addr("127.0.0.8091"), func(reqName string) bool {
		return strings.HasPrefix(reqName, "Kzz")
	})
	barProxy := service.proxyRules[len(service.proxyRules)-1]
	_, matched = barProxy.Do(context.Background(), &grpccomm.CommRequest{
		ReqSercice:   "Kzz",
		ReqInterface: "NotMatchRequest",
	})
	if matched {
		t.Errorf("barProxy.Do NotMatchRequest not match proxy rule,but got matched")
	}

	_, matched = barProxy.Do(context.Background(), &grpccomm.CommRequest{
		ReqSercice:   "Kzz",
		ReqInterface: "KzzMatchRequest",
	})
	if !matched {
		t.Errorf("barProxy.Do KzzMatchRequest expected match proxy rule,but got not match")
		return
	}

	group.ProxyService("Bar", func(ctx context.Context, req *grpccomm.CommRequest) (grpcinvoke.Service, error) {
		return grpcinvoke.Addr("service-bar"), nil
	})

	group.ProxyService("Foo", DynamicProxyTarget(func(ctx context.Context, req *grpccomm.CommRequest) (grpcinvoke.Service, error) {
		return grpcinvoke.Addr("service-foo"), nil
	}))

	if len(service.proxyRules) != 9 {
		t.Errorf("expect 9 proxy rule,got %v", len(service.proxyRules))
		return
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Errorf("ProxyInterface expect error when pass unsupport target")
			}
		}()
		group.ProxyInterface(&TestUnsupportTarget{}, &testproto.DepositRequest{})
	}()

	func() {
		defer func() {
			if recover() == nil {
				t.Errorf("ProxyService expect error when pass unsupport target")
			}
		}()
		group.ProxyService("Bridge", &TestUnsupportTarget{}, &testproto.DepositRequest{})
	}()
}
