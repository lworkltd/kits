package main

import (
	"fmt"
	"runtime"

	"github.com/Sirupsen/logrus"

	"github.com/lworkltd/kits/service/grpcinvoke"
	_ "github.com/lworkltd/kits/service/grpcinvoke/invokeimpl"
	"github.com/lworkltd/kits/service/grpcsrv/example/testproto"
	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
)

func main() {
	runtime.GOMAXPROCS(2)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
	})
	var (
		err error
	)
	req1 := &testproto.DepositRequest{}

	err = grpcinvoke.Addr("127.0.0.1:8090").Unary().Header(&testproto.AccountHeader{
		Account:  "abc",
		Password: "123",
	}).Body(req1).Response(nil)
	if err != nil {
		fmt.Println("Error", err)
		return
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

}
