package invokeimpl

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/golang/protobuf/proto"
	"github.com/lworkltd/kits/service/grpcinvoke"
	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
	"github.com/lworkltd/kits/service/restful/code"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type grpcClient struct {
	err error

	callName    string
	body        proto.Message
	header      proto.Message
	fallback    func(error) error
	serviceId   string
	serviceName string
	reqService  string
	conn        *grpc.ClientConn

	hystrixInfo hystrix.CommandConfig
	useTracing  bool
	useCircuit  bool
	ctx         context.Context

	freeConnAfterUsed bool
}

func newErrorGrpcClient(err error) *grpcClient {
	return &grpcClient{err: err}
}

func (client *grpcClient) Header(reqHeader proto.Message) grpcinvoke.Client {
	client.header = reqHeader
	return client
}

func (client *grpcClient) ReqService(reqService string) grpcinvoke.Client {
	client.reqService = reqService

	return client
}

func (client *grpcClient) Body(reqBody proto.Message) grpcinvoke.Client {
	client.body = reqBody

	if client.callName == "" {
		t := reflect.TypeOf(reqBody)
		client.callName = t.Elem().Name()
	}

	return client
}

func (client *grpcClient) Fallback(f func(error) error) grpcinvoke.Client {
	client.fallback = f
	return client
}

func buildGrpcCommRequest(client *grpcClient) (*grpccomm.CommRequest, error) {
	in := &grpccomm.CommRequest{
		ReqInterface: client.callName,
	}

	if client.body != nil {
		bodyBytes, err := proto.Marshal(client.body)
		if err != nil {
			return nil, fmt.Errorf("bad request body")
		}
		in.Body = bodyBytes
	}

	if client.header != nil {
		headerBytes, err := proto.Marshal(client.header)
		if err != nil {
			return nil, fmt.Errorf("bad grpc header")
		}
		in.Header = headerBytes
	}

	reqService := client.reqService
	if reqService == "" {
		reqService = client.serviceName
	}
	in.ReqSercice = reqService

	if client.ctx == nil {
		client.ctx = context.Background()
	}

	return in, nil
}

func (client *grpcClient) CommRequest(in *grpccomm.CommRequest) *grpccomm.CommResponse {
	if client.err != nil {
		return &grpccomm.CommResponse{
			Result:  false,
			Mcode:   "INVOKE_FAILED",
			Message: fmt.Sprintf("invoke %v failed,%v", client.serviceName, client.err),
		}
	}

	if client.freeConnAfterUsed {
		if client.conn != nil {
			defer client.conn.Close()
		}
	}

	if client.ctx == nil {
		client.ctx = context.Background()
	}

	var (
		rsp *grpccomm.CommResponse
		err error
	)

	grpcClient := grpccomm.NewCommServiceClient(client.conn)
	if !client.useCircuit {
		rsp, err = grpcClient.RpcRequest(client.ctx, in)
	} else {
		client.updateHystrix()
		var cancel context.CancelFunc
		newCtx, cancel := context.WithCancel(client.ctx)
		err = hystrix.Do(client.hytrixCommand(), func() error {
			insideRsp, insideErr := grpcClient.RpcRequest(newCtx, in)
			rsp = insideRsp
			return insideErr
		}, client.fallback)
		if nil != err && nil != cancel {
			cancel()
		}
	}

	if err != nil {
		return &grpccomm.CommResponse{
			Result:  false,
			Mcode:   "INVOKE_FAILED",
			Message: fmt.Sprintf("invoke %v failed,%v", client.serviceName, client.err),
		}
	}

	return rsp
}

func (client *grpcClient) Context(ctx context.Context) grpcinvoke.Client {
	client.ctx = ctx
	return client
}

func (client *grpcClient) Response(out proto.Message) error {
	if client.err != nil {
		return client.err
	}

	if client.freeConnAfterUsed {
		if client.conn != nil {
			defer client.conn.Close()
		}
	}

	in, err := buildGrpcCommRequest(client)
	if err != nil {
		return err
	}

	var (
		rsp *grpccomm.CommResponse
	)
	grpcClient := grpccomm.NewCommServiceClient(client.conn)
	if !client.useCircuit {
		rsp, err = grpcClient.RpcRequest(client.ctx, in)
	} else {
		client.updateHystrix()
		var cancel context.CancelFunc
		newCtx, cancel := context.WithCancel(client.ctx)
		err = hystrix.Do(client.hytrixCommand(), func() error {
			insideRsp, insideErr := grpcClient.RpcRequest(newCtx, in)
			rsp = insideRsp
			return insideErr
		}, client.fallback)
		if nil != err && nil != cancel {
			cancel()
		}
	}

	if err != nil {
		stat, ok := status.FromError(err)
		if ok {
			switch stat.Code {
			default:
				return code.NewMcode("GRPC_ERROR", err.Error())
			}
		}

		if strings.Index(err.Error(), "hystrix: timeout") >= 0 {
			return code.NewMcode("GRPC_TIMEOUT", err.Error())
		}

		return code.NewMcode("GRPC_ERROR", err.Error())
	}

	if rsp.Mcode != "" {
		return code.NewMcode(rsp.Mcode, rsp.Message)
	}

	if out != nil {
		err := proto.Unmarshal(rsp.Body, out)
		if err != nil {
			return code.NewMcode("INVOKE_BAD_GRPC_BODY", "bad grpc response body")
		}
	}

	return nil
}

func (client *grpcClient) hytrixCommand() string {
	return fmt.Sprintf("%s/%s", client.serviceName, client.callName)
}

func (client *grpcClient) updateHystrix() {
	if client.useCircuit {
		command := client.hytrixCommand()
		if "DEFAULT" != command {
			hystrix.ConfigureCommand(command, client.hystrixInfo)
		}
	}
}

// UseCircuit 启用熔断
func (client *grpcClient) UseCircuit(enable bool) grpcinvoke.Client {
	client.useCircuit = enable
	return client
}

// MaxConcurrent 最大并发请求
func (client *grpcClient) MaxConcurrent(maxConn int) grpcinvoke.Client {
	if maxConn < 30 {
		maxConn = 30
	} else if maxConn > 10000 {
		maxConn = 10000
	}
	client.hystrixInfo.MaxConcurrentRequests = maxConn
	return client
}

// Timeout 请求超时立即返回时间
func (client *grpcClient) Timeout(timeout time.Duration) grpcinvoke.Client {
	if timeout < 10*time.Millisecond {
		timeout = time.Millisecond * 10
	} else if timeout > 10000*time.Millisecond {
		timeout = 10000 * time.Millisecond
	}

	client.hystrixInfo.Timeout = int(timeout / time.Millisecond)

	return client
}

// PercentThreshold 最大错误容限
func (client *grpcClient) PercentThreshold(thresholdPercent int) grpcinvoke.Client {
	if thresholdPercent < 5 {
		thresholdPercent = 5
	} else if thresholdPercent > 100 {
		thresholdPercent = 100
	}

	client.hystrixInfo.RequestVolumeThreshold = thresholdPercent

	return client
}
