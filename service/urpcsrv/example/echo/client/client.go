package main

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/lworkltd/kits/service/urpcinvoke"
	"github.com/lworkltd/kits/service/urpcsrv/example/testpb"
	"github.com/lworkltd/kits/service/urpcsrv/urpccomm"
)

func perr(str string, err error) {
	panic(str + ": " + err.Error())
}

func MustMashalProto(msg proto.Message) []byte {
	b, err := proto.Marshal(msg)
	if err != nil {
		perr("proto.Mashal", err)
	}
	return b
}

func echoRequest(str string) []byte {
	return MustMashalProto(&urpccomm.CommRequest{
		ReqInterface: "echo",
		Body: MustMashalProto(&testpb.EchoRequest{
			Str: str,
		}),
	})
}

func main() {
	urpcinvoke.Init(&urpcinvoke.Option{
		Discover: func(string) ([]string, []string, error) {
			return []string{"localhost:8070"}, []string{"localhost:8070"}, nil
		},
	})
	rsp := &testpb.EchoResponse{}
	err := urpcinvoke.Name("myService").
		Call("echo").
		Timeout(time.Second).
		UseCircuit(true).
		Body(&testpb.EchoRequest{
			Str: "hello world",
		}).Response(rsp)
	if err != nil {
		perr("failed", err)
	}

	fmt.Println(rsp.Str)
}
