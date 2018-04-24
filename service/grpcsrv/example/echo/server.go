package main

import (
	"github.com/lworkltd/kits/service/grpcsrv"
	"github.com/lworkltd/kits/service/grpcsrv/example/echo/pb"
)

func main() {
	grpcsrv.Register("Echo", func(req *pb.EchoRequest) (*pb.EchoResponse, error) {
		return &pb.EchoResponse{
			Str: req.Str,
		}, nil
	})

	grpcsrv.Run(":8090", "ECHO_SERVER_")
}
