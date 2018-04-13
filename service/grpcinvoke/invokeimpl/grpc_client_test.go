package invokeimpl

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/lworkltd/kits/service/grpcinvoke"
	"github.com/lworkltd/kits/service/grpcsrv"
	"github.com/lworkltd/kits/service/grpcsrv/example/testproto"
	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
	"github.com/lworkltd/kits/service/restful/code"
)

func init() {
	go func() {
		grpcsrv.Register("DepositRequest", func() {})
		grpcsrv.Register("AddRequest", func(req *testproto.AddRequest) (*testproto.AddResponse, error) {
			return &testproto.AddResponse{
				Sum: req.A + req.B,
			}, nil
		})
		grpcsrv.Register("TimeoutRequest", func() { time.Sleep(time.Second * 60) })
		grpcsrv.Register("ErrorReqeust", func() error { return code.New(400, "error response") })
		grpcsrv.Run("0.0.0.0:8090", "TEST_ERROR_")
	}()
	time.Sleep(100 * time.Millisecond)
}
func TestGrpcClient(t *testing.T) {
	type args struct {
		f func(error) error
	}

	service := grpcinvoke.Addr("127.0.0.1:8090").Timeout(time.Second)
	notExistService := grpcinvoke.Addr("127.0.0.1:8092").Timeout(time.Millisecond * 100)
	tests := []struct {
		name      string
		getClient func() grpcinvoke.Client
		rsp       proto.Message
		checkRsp  func(err error, rsp proto.Message) bool
	}{
		{
			name: "normal",
			getClient: func() grpcinvoke.Client {
				return service.Grpc("DepositRequest").
					ReqService("Foo").
					Fallback(func(err error) error { return err }).
					Body(&testproto.AddRequest{}).
					Header(&testproto.AccountHeader{}).
					Timeout(time.Second).
					MaxConcurrent(10).
					PercentThreshold(20).
					UseCircuit(true).
					Context(context.Background())
			},
			rsp: nil,
			checkRsp: func(err error, rsp proto.Message) bool {
				if err != nil {
					//t.Errorf("unexpect error,got err=%v", err)
					return false
				}
				return true
			},
		},
		{
			name: "service not exist",
			getClient: func() grpcinvoke.Client {
				return notExistService.Grpc("DepositRequest").
					UseCircuit(true)
			},
			rsp: nil,
			checkRsp: func(err error, rsp proto.Message) bool {
				if err == nil {
					t.Errorf("unexpect error,got nil")
					return false
				}
				return true
			},
		},
		{
			name: "request timeout",
			getClient: func() grpcinvoke.Client {
				return service.Grpc("TimeoutRequest").
					UseCircuit(true)
			},
			rsp: nil,
			checkRsp: func(err error, rsp proto.Message) bool {
				if err == nil {
					t.Errorf("unexpect timout error,got nil")
					return false
				}
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := tt.getClient()
			err := cli.Response(tt.rsp)
			tt.checkRsp(err, tt.rsp)
		})
	}
}

func protoBytesOf(msg proto.Message) []byte {
	b, err := proto.Marshal(msg)
	if err != nil {
		panic("can not marshal msg into bytes " + err.Error())
	}
	return b
}

func TestGrpcCommReq(t *testing.T) {
	type args struct {
		f func(error) error
	}
	service := grpcinvoke.Addr("127.0.0.1:8090").Timeout(time.Second)
	tests := []struct {
		name      string
		getClient func() grpcinvoke.Client
		rsp       *grpccomm.CommResponse
		checkRsp  func(rsp *grpccomm.CommResponse) bool
		commReq   *grpccomm.CommRequest
	}{
		{
			name: "normal ok",
			getClient: func() grpcinvoke.Client {
				return service.Grpc("").UseCircuit(true)
			},
			commReq: &grpccomm.CommRequest{
				ReqInterface: "AddRequest",
				Body: protoBytesOf(&testproto.AddRequest{
					A: 100,
					B: 200,
				}),
			},
			rsp: nil,
			checkRsp: func(rsp *grpccomm.CommResponse) bool {
				if !rsp.Result {
					t.Errorf("AddRequest unexpect error got %v %v", rsp.Mcode, rsp.Message)
					return false
				}
				ret := &testproto.AddResponse{}
				if err := proto.Unmarshal(rsp.Body, ret); err != nil {
					t.Errorf("proto.Unmarshal(rsp) unexpect err,got %v", err)
					return false
				}
				if ret.Sum != 300 {
					t.Errorf("AddRequest expect sum=300，got %v", ret.Sum)
					return false
				}

				return false
			},
		},

		{
			name: "normal timeout",
			getClient: func() grpcinvoke.Client {
				return service.Grpc("").UseCircuit(true)
			},
			commReq: &grpccomm.CommRequest{
				ReqInterface: "TimeoutRequest",
			},
			rsp: nil,
			checkRsp: func(rsp *grpccomm.CommResponse) bool {
				return false
			},
		},

		{
			name: "error client",
			getClient: func() grpcinvoke.Client {
				return newErrorGrpcClient(code.New(-1, "error client"))
			},
			commReq: &grpccomm.CommRequest{},
			rsp:     nil,
			checkRsp: func(rsp *grpccomm.CommResponse) bool {
				if rsp.Result == true {
					t.Errorf("newErrorGrpcClient expect error")
					return false
				}
				return false
			},
		},
		{
			name: "error response",
			getClient: func() grpcinvoke.Client {
				return service.Grpc("").
					ReqService("Foo").UseCircuit(true)
			},
			commReq: &grpccomm.CommRequest{
				ReqInterface: "ErrorRequest",
			},
			rsp: nil,
			checkRsp: func(rsp *grpccomm.CommResponse) bool {
				if rsp.Result == true {
					t.Errorf("newErrorGrpcClient expect error")
					return false
				}
				return false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := tt.getClient()
			rsp := cli.CommRequest(tt.commReq)
			tt.checkRsp(rsp)
		})
	}
}
