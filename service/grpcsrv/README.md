grpcsrv
--------------------

- 支持灵活的注册函数，依靠消息的后缀`-Request`,`-Header`和`Context`来识别消息解析类型；
- 支持对接口进行分组，对组进行消息预处理；
- 支持与其他Grpc服务，进行同网络端口监听；
- 支持服务钩子，源生支持异常恢复，日志，防雪崩，监控上报，当然，也可以自定义钩子传入。
- 支持健康检测，参见[测试Demo](example/health/health.go)


快速开始
-------
```bash
go get github.com/lworkltd/kits
cd $GOPATH/github.com/lworkltd/kits/service/grpcsrv/example/echo
go run server.go
# open another cmd console
go run client.go
# -------------------
# Output: Hello world
```
##### [服务端代码](example/echo/server.go)

```golang
grpcsrv.Register("Echo", func(req *pb.EchoRequest) (*pb.EchoResponse, error) {
    return &pb.EchoResponse{
        Str: req.Str,
    }, nil
})
grpcsrv.Run(":8090","ECHO_SERVER_")
```

##### [客户端代码](example/echo/client.go)
```golang
rsp := pb.EchoResponse{}
grpcinvoke.Addr("127.0.0.1:8090").Unary("Echo").Body(&pb.EchoRequest{
    Str:"Hello world",
}).Response(&rsp)

fmt.Println(rsp.Str)
```


参数规则
---
Handler的规则可能是复杂的（对于想知道支持哪些格式的人），当然也是最灵活的，你可以根据自己的需要，定制适合自己的Handler，规则可以描述为如下：
```
func ([context.Context][,*XxxxRequest],[*XxxxHeader])([XxxxResponse][,error])
func ([context.Context][,*grpccomm.CommRequest) *grpccomm.CommResponse
```

> **注意：** 错误的格式会导致触发 `Panic`

##### 入参
按名称匹配，参数类型与顺序无关，且所有入参类型都不是必填的。

- Req或Request后缀（`*grpccomm.CommRequest`除外）会主动去解析[消息体](grpccomm/grpc_comm.proto#61)的数据；
- Header后缀的会主动去解析[头部](grpccomm/grpc_comm.proto#60)的数据；
- `context.Context`会透传上来；
- 如果你使用了`*grpccomm.CommRequest`作为入参，那么底层的数据，将会通过预处理后，透传上来。

> **注意：** 如果签名中不包含某种入参，则该入参则不会解析（性能相关请划重点）

```golang
func ()(...)
func (context.Context)(...)
func (*pb.XxxxRequest)(...)
func (*pb.XxxxHeader)(...)
func (context.Context, *pb.XxxxRequest)(...)
func (context.Context, *pb.XxxxRequest, *pb.XxxxHeader)(...)
func (context.Context, *pb.XxxxRequest, *pb.XxxxHeader)(...)
func (*grpccomm.CommRequest)(...)
func (context.Context, *grpccomm.CommRequest)(...)
```
##### 出参
支持4种定义的出参方式：

- 什么都不填，则一个完全正确的`*grpccomm.CommResponse`将会返回给客户端；
- 仅error，注意error强烈建议是code.Error类型，这样会返回一个对应错误码和描述信息的响应；
- `仅*grpccomm.CommResponse`,将原封不动的透传给客户端；
- `(*pb.XxxxResponse,error)`，注意顺序必须`*pb.XxxxResponse`在前，`error`在后（这是没有特殊原因的，仅为了更加符合编程规范而已）。

```golang
func (...) 
func (...) error
func (...) (*pb.XxxxResponse, error) 
func (...) (*grpccomm.CommResponse)
```

例子
-----
#### 典型的应用场景

##### 服务端

```golang
grpcsrv.Register(&testproto.AddRequest{}, func Add(req *testproto.AddRequest) (*testproto.AddResponse, error) {
    return &testproto.AddResponse{
        Sum: req.A + req.B,
    }, nil
})
grpcsrv.Run("0.0.0.0:8090", "TESTECHO_")
```

##### 客户端

```golang
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

```golang
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

```golang
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

```golang
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
```golang
// 应用CheckAcount
authGroup := grpcsrv.Group("auth", CheckAcount)
authGroup.Register("LoginRequest",Login)

// 首先应用CheckAcount，再应用CheckBalance
balanceGroup := authGroup.Group("balance",CheckBalance)
balanceGroup.Register("DepositRequest",Deposit)
balanceGroup.Register("WithdrawRequest",Withdraw)
grpcsrv.Run("0.0.0.0:8090", "TEST_")
```