package main

import (
	"context"
	"net/http"

	_ "net/http/pprof"

	"github.com/lworkltd/kits/service/urpcsrv"
	"github.com/lworkltd/kits/service/urpcsrv/example/testpb"
	"github.com/lworkltd/kits/service/urpcsrv/urpccomm"
)

func preCheck(context.Context, *urpccomm.CommRequest) error {
	return nil
}

func Echo(req *testpb.EchoRequest) (*testpb.EchoResponse, error) {
	return &testpb.EchoResponse{
		Str: req.Str,
	}, nil
}

func Add(req *testpb.AddRequest) (*testpb.AddResponse, error) {
	return nil, nil
}

func main() {
	urpcsrv.UseHook(
	// 向监控上报请求结果信息
	//urpcsrv.HookReportMonitor(&report.MonitorReporter{}),
	// 打印日志
	//urpcsrv.HookLogger,
	// 防止雪崩
	//urpcsrv.HookDefenceSlowSide(2000),
	// 异常恢复
	//urpcsrv.HookRecovery,
	)

	go http.ListenAndServe(":8080", nil)
	urpcsrv.Register("echo", Echo)
	aGroup := urpcsrv.Group("aGroup", preCheck)
	aGroup.Register("add", Add)
	urpcsrv.Run(":8070", "TEST_SERVICE")
}
