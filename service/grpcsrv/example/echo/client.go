package main

import (
	"fmt"

	"github.com/lworkltd/kits/service/grpcinvoke"
	_ "github.com/lworkltd/kits/service/grpcinvoke/invokeimpl"
	"github.com/lworkltd/kits/service/grpcsrv/example/echo/pb"
)

func main() {
	rsp := pb.EchoResponse{}
	grpcinvoke.Addr("127.0.0.1:8090").Unary("Echo").Body(&pb.EchoRequest{
		Str: "Hello world",
	}).Response(&rsp)

	fmt.Println(rsp.Str)
}
