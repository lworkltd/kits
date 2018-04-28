package main

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/service/grpcinvoke"
	_ "github.com/lworkltd/kits/service/grpcinvoke/invokeimpl"
	"github.com/lworkltd/kits/service/grpcsrv/example/echo/pb"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{DisableColors: true})
	rsp := pb.EchoResponse{}
	logrus.SetLevel(logrus.DebugLevel)
	grpcinvoke.Init(&grpcinvoke.Option{
		Discover: func(n string) ([]string, []string, error) {
			if n == "MyService" {
				return []string{"127.0.0.1:8090"}, []string{"127.0.0.1:8090"}, nil
			}

			return nil, nil, fmt.Errorf("service not found")
		},
	})

	err := grpcinvoke.Name("MyService").Unary("Echo").Body(&pb.EchoRequest{
		Str: "Hello world",
	}).DoLogger(true).Response(&rsp)

	if err != nil {
		fmt.Println("Echo err", err)
		return
	}

	fmt.Println(rsp.Str)
}
