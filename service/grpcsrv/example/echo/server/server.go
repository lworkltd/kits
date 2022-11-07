package main

import (
	"fmt"

	"github.com/lworkltd/kits/service/grpcsrv"
	"github.com/lworkltd/kits/service/grpcsrv/example/echo/pb"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{DisableColors: true})
	grpcsrv.UseHook(grpcsrv.HookLogger)
	grpcsrv.Register("Echo", func(req *pb.EchoRequest) (*pb.EchoResponse, error) {
		return &pb.EchoResponse{
			Str: req.Str,
		}, fmt.Errorf("error")
	})

	grpcsrv.Run(":8090", "ECHO_SERVER")
}
