package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/lworkltd/kits/service/grpcinvoke"
	_ "github.com/lworkltd/kits/service/grpcinvoke/invokeimpl"
	"github.com/lworkltd/kits/service/grpcsrv/example/testproto"
	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
	"github.com/lworkltd/kits/service/restful/code"
)

func main() {
	runtime.GOMAXPROCS(2)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
	})
	var (
		err error
	)

	// 服务发现函数
	discovery := func(name string) ([]string, []string, error) {
		if name == "MyService" {
			return []string{"127.0.0.1:8090"}, []string{"myservice-1"}, nil
		}

		return nil, nil, fmt.Errorf("service %s not found", name)
	}

	// 初始化
	grpcinvoke.Init(&grpcinvoke.Option{
		Discover:                     discovery,
		UseCircuit:                   true,
		DefaultTimeout:               time.Second * 10,
		DefaultErrorPercentThreshold: 25,
		DefaultMaxConcurrentRequests: 50,
	})

	req1 := &testproto.DepositRequest{}

	err = grpcinvoke.Name("MyService").Unary().Header(&testproto.AccountHeader{
		Account:  "abc",
		Password: "123",
	}).Body(req1).Response(nil)
	if err != nil {
		fmt.Println("Error", err)
		//return
	}

	req2 := &testproto.CalculateStrLenRequest{}

	err = grpcinvoke.Addr("127.0.0.1:8090").Unary().Header(&grpccomm.CommHeader{}).Body(req2).Response(nil)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	req3 := &testproto.AddRequest{
		A: 1,
		B: -2,
	}
	rsp3 := &testproto.AddResponse{}
	err = grpcinvoke.Addr("127.0.0.1:8090").Unary().Header(&grpccomm.CommHeader{}).Body(req3).Response(rsp3)
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	fmt.Println("AddResponse", rsp3.Sum)

	req4 := &testproto.AddRequest{
		A: 1,
		B: -2,
	}
	rsp4 := &testproto.AddResponse{}
	err = grpcinvoke.Addr("127.0.0.1:8090").Unary("Agent").Header(&grpccomm.CommHeader{}).Body(req4).Response(rsp4)
	if err != nil {
		fmt.Println("Error Agent", err)
		return
	}
	fmt.Println("AddResponse", rsp4.Sum)

	rsp5 := &testproto.DeleteUserRequest{}
	err = grpcinvoke.Addr("127.0.0.1:8090").Unary().Header(&grpccomm.CommHeader{
		BaseInfo: &grpccomm.BaseInfo{
			TenantId:  "T001234",
			XApiToken: "abc",
		},
	}).Body(rsp5).Response(nil)

	if err != nil {
		fmt.Println("Error DeleteUserRequest", err)
		return
	}

	rsp6 := &testproto.ReportRequest{}
	err = grpcinvoke.Addr("127.0.0.1:8090").Unary().Header(&grpccomm.CommHeader{}).Body(rsp6).Response(nil)
	if err != nil {
		fmt.Println("Error ReportRequest", err)
		return
	}

	rsp7 := &testproto.ReportRequest{}
	err = grpcinvoke.Addr("127.0.0.1:8090").Unary("Agent").Header(&grpccomm.CommHeader{}).Body(rsp7).Response(nil)
	if err != nil {
		fmt.Println("Error Agent", err)
		return
	}

	err = grpcinvoke.Addr("127.0.0.1:8090").Unary("MustPanic").Response(nil)
	if err == nil {
		fmt.Println("Expect panic error")
		return
	}
	fmt.Println("Got expeted error", err.(code.Error).Mcode())

	err = grpcinvoke.Name("MyService").
		Unary("HighDelay").
		Body(&testproto.HighDelayRequest{
			DelaySeconds: 2,
		}).UseCircuit(true).
		Timeout(time.Second * 1).
		Response(nil)
	if err == nil {
		fmt.Println("Expect timeout error,but got ok")
		return
	}
	fmt.Println("Got expeted error", err.Error())

	err = grpcinvoke.Name("MyService").
		Unary("HighDelay").
		Body(&testproto.HighDelayRequest{
			DelaySeconds: 2,
		}).UseCircuit(true).
		Timeout(time.Second * 3).
		Response(nil)
	if err != nil {
		fmt.Println("unexpect error", err.Error())
		return
	}

}
