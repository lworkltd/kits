package bench

import (
	"testing"
	"time"

	"github.com/lworkltd/kits/service/urpcinvoke"
	"github.com/lworkltd/kits/service/urpcsrv/example/testpb"
)

func testEcho() {
	rsp := &testpb.EchoResponse{}
	err := urpcinvoke.Addr("localhost:8070").
		Call("echo").
		Timeout(time.Second).
		UseCircuit(true).
		Body(&testpb.EchoRequest{
			Str: "hello world",
		}).Response(rsp)
	if err != nil {
		panic("failed" + err.Error())
	}
}

func BenchmarkEcho(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testEcho()
	}
}
