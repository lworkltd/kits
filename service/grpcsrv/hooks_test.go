package grpcsrv

import (
	"testing"
	"time"

	context "golang.org/x/net/context"

	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
	"github.com/lworkltd/kits/service/restful/code"
)

func TestHookRecover(t *testing.T) {
	type args struct {
		f HandlerFunc
	}

	mcodePrefix = "ERROR"
	tests := []struct {
		name      string
		args      args
		wantMcode string
	}{
		{
			name: "unkown panic",
			args: args{
				f: func(ctx context.Context, commReq *grpccomm.CommRequest) (commRsp *grpccomm.CommResponse) {
					panic("unkown error")
					return nil
				},
			},
			wantMcode: "GRPC_INTERNAL_ERROR",
		},
		{
			name: "code panic",
			args: args{
				f: func(ctx context.Context, commReq *grpccomm.CommRequest) (commRsp *grpccomm.CommResponse) {
					panic(code.New(100, "code error"))
					return nil
				},
			},
			wantMcode: "ERROR_100",
		},

		{
			name: "mcode panic",
			args: args{
				f: func(ctx context.Context, commReq *grpccomm.CommRequest) (commRsp *grpccomm.CommResponse) {
					panic(code.NewMcode("NOT_FOUND", "mcode error"))
					return nil
				},
			},
			wantMcode: "NOT_FOUND",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newf := HookRecovery(tt.args.f)
			rsp := newf(context.Background(), &grpccomm.CommRequest{})
			if rsp.Result {
				t.Errorf("HookRecover want rsp.Result=false,got true")
				return
			}

			if rsp.Mcode != tt.wantMcode {
				t.Errorf("HookRecover want rsp.Mcode=%s,got %s", tt.wantMcode, rsp.Mcode)
				return
			}
		})
	}
}

func TestHookLogger(t *testing.T) {
	type args struct {
		f HandlerFunc
	}
	tests := []struct {
		name    string
		args    args
		perpare func()
	}{
		{

			name: "normal ok",
			args: args{
				f: func(ctx context.Context, commReq *grpccomm.CommRequest) (commRsp *grpccomm.CommResponse) {
					return &grpccomm.CommResponse{
						Result: true,
					}
				},
			},
			perpare: func() {
				MinWarningDelay = time.Second * 10
			},
		},
		{

			name: "error",
			args: args{
				f: func(ctx context.Context, commReq *grpccomm.CommRequest) (commRsp *grpccomm.CommResponse) {
					return &grpccomm.CommResponse{
						Result: false,
						Mcode:  "ERROR_1",
					}
				},
			},
			perpare: func() {
				MinWarningDelay = time.Second * 10
			},
		},
		{

			name: "latency warning",
			args: args{
				f: func(ctx context.Context, commReq *grpccomm.CommRequest) (commRsp *grpccomm.CommResponse) {
					time.Sleep(time.Millisecond * 100)
					return nil
				},
			},
			perpare: func() {
				MinWarningDelay = time.Millisecond * 100
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.perpare()
			gotF := HookLogger(tt.args.f)
			gotF(context.Background(), &grpccomm.CommRequest{
				ReqInterface: "TestRequest",
				ReqSercice:   "TestService",
			})
		})
	}
}

func TestDefenceSlowSideHook(t *testing.T) {
	type args struct {
		n int32
	}
	tests := []struct {
		name string
		args args
	}{

		{

			name: "latency warning",
			args: args{
				n: 100,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HookDefenceSlowSide(tt.args.n)
			f := got(func(ctx context.Context, commReq *grpccomm.CommRequest) (commRsp *grpccomm.CommResponse) {
				return nil
			})
			f(context.Background(), &grpccomm.CommRequest{})
		})
	}
}
