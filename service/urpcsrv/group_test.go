package urpcsrv

import (
	"fmt"
	"reflect"
	"testing"

	context "golang.org/x/net/context"

	_ "github.com/lworkltd/kits/service/grpcinvoke/invokeimpl"
	"github.com/lworkltd/kits/service/urpcsrv/urpccomm"
)

func TestInterfaceGroupSubGroup(t *testing.T) {
	var sum string
	service := newService()
	addabc := service.Group("add-abc", func(context.Context, *urpccomm.CommRequest) error {
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

	add123 := addabc.Group("add-123", func(context.Context, *urpccomm.CommRequest) error {
		sum += "123"
		return nil
	})

	add123x := service.Group("add-123")
	if !reflect.DeepEqual(add123, add123x) {
		t.Errorf("unexpect equal group=%v,got %v", add123, add123x)
		return
	}

	addxyz := add123.Group("add-xyz", func(context.Context, *urpccomm.CommRequest) error {
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

	resultErr := add123.Group("add-123", func(context.Context, *urpccomm.CommRequest) error {
		return fmt.Errorf("error")
	})
	err := resultErr.doPipe(nil, nil)
	if err == nil {
		t.Errorf("expect result error")
		return
	}
}
