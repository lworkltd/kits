package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/lworkltd/kits/service/grpcinvoke"
	_ "github.com/lworkltd/kits/service/grpcinvoke/invokeimpl"
	"github.com/lworkltd/kits/service/grpcsrv/example/testproto"
	"github.com/lworkltd/kits/service/version"
	"github.com/lworkltd/kits/utils/jsonize"
)

func main() {
	runtime.GOMAXPROCS(2)

	var err error
	req1 := &testproto.DepositRequest{}
	rsp := version.VersionResponse{}
	timeoutCtx, _ := context.WithTimeout(context.Background(), time.Second*10)
	err = grpcinvoke.Addr("127.0.0.1:8090").
		Unary("_AppVersion").
		Body(req1).
		Context(timeoutCtx).
		Response(&rsp)
	if err != nil {
		fmt.Println("grpcinvoke call failed", err)
		return
	}

	fmt.Println(jsonize.V(rsp, true))
}
