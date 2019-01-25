package main

import (
	"fmt"
	"time"

	context "golang.org/x/net/context"

	"net/http"
	_ "net/http/pprof"

	"github.com/golang/protobuf/proto"
	"github.com/lworkltd/kits/service/grpcsrv"
	"github.com/lworkltd/kits/service/grpcsrv/example/testproto"
	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
	"github.com/lworkltd/kits/service/grpcsrv/report"
	"github.com/lworkltd/kits/service/restful/code"
)

// InvokeTestEcho 处理函数
func InvokeTestEcho(ctx context.Context, header *grpccomm.CommHeader, req *grpccomm.InvokeTestEchoRequest) (*grpccomm.InvokeTestEchoResponse, error) {
	rsp := &grpccomm.InvokeTestEchoResponse{
		Str: req.Str,
	}

	return rsp, nil
}

// CalLen 求长度
func CalLen(header *grpccomm.CommHeader, req *testproto.CalculateStrLenRequest) (*testproto.CalculateStrLenResponse, error) {
	rsp := &testproto.CalculateStrLenResponse{
		Len: int32(len(req.Str)),
	}

	return rsp, nil
}

// Add 加
func Add(ctx context.Context, req *testproto.AddRequest) (*testproto.AddResponse, error) {
	return &testproto.AddResponse{
		Sum: req.A + req.B,
	}, nil
}

// Deposit 入金
func Deposit(ctx context.Context, req *testproto.DepositRequest) (*testproto.DepositResponse, error) {
	return &testproto.DepositResponse{
		Timestamp: time.Now().Unix(),
	}, nil
}

// DeleteUser 删除
func DeleteUser(req *testproto.DeleteUserRequest) error {
	return nil
}

// Report 加
func Report(req *testproto.ReportRequest) {
}

// Agent 反向代理，透传
func Agent(req *grpccomm.CommRequest) *grpccomm.CommResponse {
	fmt.Println(req.ReqInterface)
	return nil
}

// Ping Ping
func Ping() {

}

// MustPanic 必然crash
func MustPanic() {
	panic("it's panic")
}

// HighDelay 高延迟
func HighDelay(req *testproto.HighDelayRequest) {
	time.Sleep(time.Duration(req.DelaySeconds) * time.Second)
}

// CheckAcount 检查信息
func CheckAcount(ctx context.Context, commReq *grpccomm.CommRequest) error {
	header := &testproto.AccountHeader{}
	if err := proto.Unmarshal(commReq.GetHeader(), header); err != nil {
		return code.NewMcode("BAD_HEADER", "auth failed")
	}

	if header.Account != "abc" && header.Password != "123" {
		return code.NewMcode("BAD_ACCOUNT", "auth failed")
	}

	fmt.Println("CheckAmount")

	return nil
}

// TenantCheck 租户检查
func TenantCheck(ctx context.Context, commReq *grpccomm.CommRequest) error {
	header := &grpccomm.CommHeader{}
	if err := proto.Unmarshal(commReq.GetHeader(), header); err != nil {
		return code.NewMcode("BAD_HEADER", "auth failed")
	}

	if header.BaseInfo == nil {
		return code.NewMcode("BAD_TENANT_AUTH", "lack tenant auth info")
	}

	if header.BaseInfo.TenantId == "T001234" && header.BaseInfo.XApiToken != "abc" {
		return code.NewMcode("BAD_TENANT_AUTH", "auth tenant failed")
	}

	fmt.Println("CheckAmount")

	return nil
}

func main() {

	// 当接收到消息时会首先进入钩子，钩子的顺序为在FILO
	// 例如： [hook1,hook2,hook3,hook4]
	// hook1 {
	//		hook2 {
	//			hook3 {
	//				hook4 {
	//					处理函数
	//				}
	//			}
	//	 	}
	// }
	//
	// 注意：放在HookRecover之后的钩子，会可能因为panic，被跳过处理。

	grpcsrv.UseHook(
		// 向监控上报请求结果信息
		grpcsrv.HookReportMonitor(&report.MonitorReporter{}),
		// 打印日志
		grpcsrv.HookLogger,
		// 防止雪崩
		grpcsrv.HookDefenceSlowSide(2000),
		// 异常恢复
		grpcsrv.HookRecovery,
	)

	// 分组用于预处理
	// 所有该组下的请求，都会首先通过此函数，如果此函数返回错误，则不进一步处理
	authGroup := grpcsrv.Group("auth", CheckAcount)
	// 标准头 + 标准处理函数
	grpcsrv.Register(&testproto.CalculateStrLenRequest{}, CalLen)
	// 标准头 + 无Context处理函数
	authGroup.Register(&testproto.DepositRequest{}, Deposit)
	// 不包含头
	grpcsrv.Register(&testproto.AddRequest{}, Add)
	// 仅有入参的处理函数
	grpcsrv.Register(&testproto.ReportRequest{}, Report)
	tenantAuthGroup := grpcsrv.Group("tenantCheck", TenantCheck)
	// 仅有入参,头部和返回错误的处理函数
	tenantAuthGroup.Register(&testproto.DeleteUserRequest{}, DeleteUser)
	// 测试全透传模式(可用于路由)
	grpcsrv.Register("Agent", Agent)
	// 使用字符串作为接口名称
	grpcsrv.Register("Ping", Ping)
	// 导致panic
	grpcsrv.Register("MustPanic", MustPanic)
	// 模拟高延迟服务
	grpcsrv.Register("HighDelay", HighDelay)

	// For pprof
	go func() {
		http.Handle("/grpcsrv", grpcsrv.DebugHttpHandler())
		http.ListenAndServe(":8080", nil)
	}()

	// 后端grpc服务监听
	// grpcsrv.Run("0.0.0.0:8090", "TESTECHO_")

	// 启动WebServer
	// grpcsrv.RunWeb("0.0.0.0:8090", "TESTECHO_")

	// 启动WebServer TLS
	grpcsrv.RunWebTLS("0.0.0.0:8090", "TESTECHO_", "lwork.crt", "lwork.key")
}
