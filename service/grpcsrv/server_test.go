package grpcsrv

import (
	"fmt"
	"reflect"
	"testing"

	context "golang.org/x/net/context"

	"github.com/golang/protobuf/proto"
	"github.com/lworkltd/kits/service/grpcsrv/example/testproto"
	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
	"google.golang.org/grpc"
)

func protoBytesOf(msg proto.Message) []byte {
	b, err := proto.Marshal(msg)
	if err != nil {
		panic("can not marshal msg into bytes " + err.Error())
	}
	return b
}

func TestServiceRpcRequest(t *testing.T) {
	type args struct {
		ctx     context.Context
		commReq *grpccomm.CommRequest
	}
	tests := []struct {
		name    string
		service *Service
		args    args
		want    *grpccomm.CommResponse
		wantErr bool
	}{
		{
			name: "normal request",
			service: func() *Service {
				Register("DepositRequest", func(ctx context.Context, header *testproto.AccountHeader, req *testproto.DepositRequest) (*testproto.DepositResponse, error) {
					if req.Money != -100 {
						panic("request.Money expect -100")
					}
					if header.Account != "xiaoming" {
						panic("header.Account expect xiaoming")
					}

					return &testproto.DepositResponse{
						Timestamp: 123,
					}, nil
				})
				return defaultService
			}(),
			args: args{
				ctx: context.Background(),
				commReq: &grpccomm.CommRequest{
					ReqInterface: "DepositRequest",
					Body: protoBytesOf(&testproto.DepositRequest{
						Money: -100,
					}),
					Header: protoBytesOf(&testproto.AccountHeader{
						Account: "xiaoming",
					}),
				},
			},
			want: &grpccomm.CommResponse{
				Result: true,
				Body: protoBytesOf(&testproto.DepositResponse{
					Timestamp: 123,
				}),
			},
		},

		{
			name: "new service",
			service: func() *Service {
				service := newService()
				service.Register("DepositRequest", func() {})
				return service
			}(),
			args: args{
				ctx: context.Background(),
				commReq: &grpccomm.CommRequest{
					ReqInterface: "DepositRequest",
				},
			},
			want: &grpccomm.CommResponse{
				Result: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(tt.service.methodIndex)
			got, err := tt.service.RpcRequest(tt.args.ctx, tt.args.commReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.RpcRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.RpcRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	type args struct {
		host      string
		errPrefix string
		grpcOpts  []grpc.ServerOption
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				host: ":70000",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Run(tt.args.host, tt.args.errPrefix, tt.args.grpcOpts...); (err != nil) != tt.wantErr {
				t.Errorf("ListenAndServe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUseHook(t *testing.T) {
	type args struct {
		hooks []HookFunc
	}
	tests := []struct {
		name string
		args args
	}{
		{
			args: args{
				hooks: []HookFunc{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UseHook(tt.args.hooks...)
		})
	}
}
