grpcsrv
--------------------

- 支持灵活的注册函数，依靠消息的后缀Request,Header，Context来识别消息结构
- 支持对接口进行分组，对组进行消息预处理
- 支持与其他Grpc服务，进行同网络端口监听
- 支持服务钩子，源生支持异常恢复，日志，防雪崩，监控上报，当然，也可以自定义钩子传入

例子
-----
#### 典型的应用场景
##### 服务端
```
grpcsrv.Register(&testproto.AddRequest{}, func Add(req *testproto.AddRequest) (*testproto.AddResponse, error) {
    return &testproto.AddResponse{
        Sum: req.A + req.B,
    }, nil
})
grpcsrv.Run("0.0.0.0:8090", "TESTECHO_")
```

##### 客户端
```
req := &testproto.AddRequest{
    A: 1,
    B: -2,
}
rsp := &testproto.AddResponse{}
err = grpcinvoke.Addr("127.0.0.1:8090").Unary("AddRequest").Body(req).Response(rsp)
if err != nil {
    fmt.Println("Error AddRequest", err)
    return
}
fmt.Println("AddResponse", rsp.Sum)
```
#### 如果需要和其他的Grpc Server共用服务器端口
##### 服务端
```
grpcServer := grpc.NewServer()

// 同步请求服务
grpcsrv.SetErrPrefix("TEST_")
grpcsrv.Register("AddRequest", Add)
grpccomm.RegisterCommServiceServer(grpcServer, grpcsrv.DefaultService())

// 双向流服务
pb.RegisterLmaxQuoteServer(grpcServer, &myService{})

lis, err := net.Listen("tcp", "127.0.0.1:8090")
if err != nil {
    panic(fmt.Errorf("failed to listen: %v", err))
}

grpcServer.Serve(lis)
```

##### 客户端
```
req := &testproto.AddRequest{
    A: 1,
    B: -2,
}
rsp := &testproto.AddResponse{}
err := grpcinvoke.Addr("127.0.0.1:8090").Unary().Body(req).Response(rsp)
if err != nil {
    fmt.Println("AddRequest Error", err)
    return
}
fmt.Println("AddResponse", rsp.Sum)

conn, err := grpc.Dial("127.0.0.1:8090", grpc.WithInsecure())
if err != nil {
    fmt.Println("Dial Error", err)
    return
}
cli := pb.NewLmaxQuoteClient(conn)
stream, err := cli.Prices(context.Background())
if err != nil {
    fmt.Println("Dial Error", err)
}
for {
    err := stream.Send(&pb.PricesRequest{
        PriceType: 0,
    })
    if err != nil {
        fmt.Println("stream.Send Error", err)
        return
    }

    r, err := stream.Recv()
    if err != nil {
        fmt.Println("stream.Recv Error", err)
        return
    }

    fmt.Println(r.Price)
}
```

#### 钩子
```
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

```

#### 预处理组
```
// 应用CheckAcount
authGroup := grpcsrv.Group("auth", CheckAcount)
authGroup.Register("LoginRequest",Login)

// 首先应用CheckAcount，再应用CheckBalance
balanceGroup := authGroup.Group("balance",CheckBalance)
balanceGroup.Register("DepositRequest",Deposit)
balanceGroup.Register("WithdrawRequest",Withdraw)
grpcsrv.Run("0.0.0.0:8090", "TESTECHO_")
```