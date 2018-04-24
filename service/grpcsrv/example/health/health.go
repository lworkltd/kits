package main

import (
	"fmt"

	"google.golang.org/grpc"

	context "golang.org/x/net/context"
	hv1 "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	c, err := grpc.Dial("127.0.0.1:8090", grpc.WithInsecure())
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}

	client := hv1.NewHealthClient(c)
	ret, err := client.Check(context.Background(), &hv1.HealthCheckRequest{
		Service: "grpccomm.CommService",
	})
	if err != nil {
		fmt.Println("client.Check err", err)
		return
	}

	fmt.Println("Service status is", ret.Status)
}
