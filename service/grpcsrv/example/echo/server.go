package main

import (
	"github.com/lworkltd/kits/service/grpcsrv"
	"github.com/lworkltd/kits/service/grpcsrv/example/echo/pb"
	"github.com/lworkltd/kits/service/restful/code"
)

func main() {
	grpcsrv.UseHook(grpcsrv.HookLogger)
	grpcsrv.Register("Echo", func(req *pb.EchoRequest) (*pb.EchoResponse, error) {
		return &pb.EchoResponse{
			Str: req.Str,
		}, code.New(1000, "")
	})

	grpcsrv.Run(":8090", "ECHO_SERVER")
}
