package grpcsrv

import (
	"reflect"
	"testing"
	"time"

	context "golang.org/x/net/context"

	"github.com/lworkltd/kits/service/grpcsrv/example/testproto"
	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
	"github.com/lworkltd/kits/service/restful/code"
)

func TestCreateRegReqName(t *testing.T) {
	type args struct {
		reqBody interface{}
		f       interface{}
	}
	tests := []struct {
		name    string
		args    args
		reqName string
	}{
		// 接口名称
		{
			name: "string/none/none",
			args: args{
				reqBody: "DepositRequest",
				f:       func() {},
			},
			reqName: "DepositRequest",
		},
		{
			name: "requestBody/none/none",
			args: args{
				reqBody: testproto.DepositRequest{},
				f:       func() {},
			},
			reqName: "DepositRequest",
		},
		{
			name: "requestBodyPtr/none/none",
			args: args{
				reqBody: &testproto.DepositRequest{},
				f:       func() {},
			},
			reqName: "DepositRequest",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regInfo := createRegInfo(tt.args.reqBody, tt.args.f)
			if regInfo.reqName != tt.reqName {
				t.Errorf("createRegInfo failed,expect reqName %v got %v", tt.reqName, regInfo.reqName)
				return
			}
		})
	}
}

type NotProtoMessageRequest struct {
}

type UnsupportRequestXXX struct {
}

type NotProtoMessageHeader struct {
}

type UnsupportHeaderXXX struct {
}

type NotProtoMessageResponse struct {
}

func TestCreateRegInput(t *testing.T) {
	type args struct {
		reqBody interface{}
		f       interface{}
	}
	tests := []struct {
		name       string
		args       args
		needErr    bool
		checkPos   bool
		ctxPos     int
		reqPos     int
		commReqPos int
		headerPos  int
	}{
		// 接口名称
		{
			name: "request must be pointer",
			args: args{
				reqBody: "Test",
				f:       func(testproto.DepositRequest) {},
			},
			needErr: true,
		},
		{
			name: "request must be pointer",
			args: args{
				reqBody: "Test",
				f:       func(testproto.DepositRequest) {},
			},
			needErr: true,
		},
		{
			name: "unkown type",
			args: args{
				reqBody: "Test",
				f:       func(*UnsupportRequestXXX) {},
			},
			needErr: true,
		},
		{
			name: "ok",
			args: args{
				reqBody: "Test",
				f:       func(*testproto.DepositRequest) {},
			},
			needErr: false,
		},
		{
			name: "not proto message request",
			args: args{
				reqBody: "Test",
				f:       func(*NotProtoMessageRequest) {},
			},
			needErr: true,
		},
		{
			name: "commReq ptr only",
			args: args{
				reqBody: "Test",
				f:       func(*grpccomm.CommRequest) {},
			},
			needErr: false,
		},
		{
			name: "commReq only error",
			args: args{
				reqBody: "Test",
				f:       func(grpccomm.CommRequest) {},
			},
			needErr: true,
		},
		{
			name: "context only",
			args: args{
				reqBody: "Test",
				f:       func(context.Context) {},
			},
			needErr: false,
		},
		{
			name: "context ptr error",
			args: args{
				reqBody: "Test",
				f:       func(*context.Context) {},
			},
			needErr: true,
		},
		{
			name: "context and request",
			args: args{
				reqBody: "Test",
				f:       func(context.Context, *testproto.DepositRequest) {},
			},
			needErr: false,
		},
		{
			name: "header ptr only",
			args: args{
				reqBody: "Test",
				f:       func(*testproto.AccountHeader) {},
			},
			needErr: false,
		},
		{
			name: "header not ptr",
			args: args{
				reqBody: "Test",
				f:       func(testproto.AccountHeader) {},
			},
			needErr: true,
		},
		{
			name: "unsupport header",
			args: args{
				reqBody: "Test",
				f:       func(*NotProtoMessageHeader) {},
			},
			needErr: true,
		},

		{
			name: "request&context&header",
			args: args{
				reqBody: "Test",
				f:       func(*testproto.DepositRequest, context.Context, *testproto.AccountHeader) {},
			},
			needErr:   false,
			checkPos:  true,
			ctxPos:    1,
			reqPos:    0,
			headerPos: 2,
		},

		{
			name: "request&context",
			args: args{
				reqBody: "Test",
				f:       func(context.Context, *testproto.DepositRequest) {},
			},
			needErr:   false,
			checkPos:  true,
			ctxPos:    0,
			reqPos:    1,
			headerPos: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.needErr != (r != nil) {
					t.Errorf("createRegInfo expect error %v got %v", tt.needErr, r)
					return
				}
			}()
			regInfo := createRegInfo(tt.args.reqBody, tt.args.f)
			if tt.checkPos {
				if regInfo.reqInIndex != tt.reqPos || regInfo.ctxInIndex != tt.ctxPos || regInfo.headerInIndex != tt.headerPos {
					t.Errorf("expect input pos ctx=%d req=%d header=%d, got ctx=%d req=%d header=%d",
						tt.ctxPos, tt.reqPos, tt.headerPos,
						regInfo.ctxInIndex, regInfo.reqInIndex, regInfo.headerInIndex,
					)
				}
			}
		})
	}
}

func TestCreateRegOutput(t *testing.T) {
	type args struct {
		reqBody interface{}
		f       interface{}
	}
	tests := []struct {
		name    string
		args    args
		needErr bool
	}{
		{
			name: "error only",
			args: args{
				reqBody: "Test",
				f:       func() error { return nil },
			},
			needErr: false,
		},
		{
			name: "commRsp only",
			args: args{
				reqBody: "Test",
				f:       func() *grpccomm.CommResponse { return nil },
			},
			needErr: false,
		},
		{
			name: "commRsp not ptr",
			args: args{
				reqBody: "Test",
				f:       func() grpccomm.CommResponse { return grpccomm.CommResponse{} },
			},
			needErr: true,
		},
		{
			name: "single output neither type of error nor type of *grpccomm.CommResponse",
			args: args{
				reqBody: "Test",
				f:       func() *UnsupportHeaderXXX { return nil },
			},
			needErr: true,
		},
		{
			name: "2 output",
			args: args{
				reqBody: "Test",
				f:       func() (*testproto.DepositResponse, error) { return nil, nil },
			},
		},
		{
			name: "2 output with first not a pointer",
			args: args{
				reqBody: "Test",
				f:       func() (testproto.DepositResponse, error) { return testproto.DepositResponse{}, nil },
			},
			needErr: true,
		},

		{
			name: "2 output with first not implemented proto.Message",
			args: args{
				reqBody: "Test",
				f:       func() (*NotProtoMessageResponse, error) { return nil, nil },
			},
			needErr: true,
		},
		{
			name: "2 output with second not type of error",
			args: args{
				reqBody: "Test",
				f:       func() (*testproto.DepositResponse, *NotProtoMessageResponse) { return nil, nil },
			},
			needErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.needErr != (r != nil) {
					t.Errorf("createRegInfo expect error %v got %v", tt.needErr, r)
					return
				}
			}()
			createRegInfo(tt.args.reqBody, tt.args.f)
		})
	}
}

func TestCall0(t *testing.T) {
	type args struct {
		ctx         context.Context
		headerValue reflect.Value
		bodyValue   reflect.Value
		commReq     *grpccomm.CommRequest
		regInfo     *RegisterInfo
	}
	bodyValue := reflect.ValueOf(&testproto.DepositRequest{
		Money: 100,
	})
	headerValue := reflect.ValueOf(&testproto.AccountHeader{
		Account: "xiaoming",
	})
	commReq := &grpccomm.CommRequest{}
	tests := []struct {
		name    string
		args    args
		wantRsp bool
		wantErr bool
	}{
		{
			name: "3 input 2 output normal",
			args: args{
				bodyValue:   bodyValue,
				headerValue: headerValue,
				commReq:     commReq,
				ctx:         context.Background(),
				regInfo: createRegInfo("test", func(ctx context.Context, h *testproto.AccountHeader, r *testproto.DepositRequest) (*testproto.DepositResponse, error) {
					if r.Money != 100 {
						panic("req body not the value want pass")
					}
					if h.Account != "xiaoming" {
						panic("req header not the value want pass")
					}
					return &testproto.DepositResponse{
						Timestamp: time.Now().Unix(),
					}, nil
				}),
			},
		},

		{
			name: "3 input 2 output nil rsp",
			args: args{
				bodyValue:   bodyValue,
				headerValue: headerValue,
				commReq:     commReq,
				ctx:         context.Background(),
				regInfo: createRegInfo("test", func(ctx context.Context, h *testproto.AccountHeader, r *testproto.DepositRequest) (*testproto.DepositResponse, error) {
					if r.Money != 100 {
						panic("req body not the value want pass")
					}
					if h.Account != "xiaoming" {
						panic("req header not the value want pass")
					}
					return nil, nil
				}),
			},
		},
		{
			name: "3 input 2 output rsp error",
			args: args{
				bodyValue:   bodyValue,
				headerValue: headerValue,
				commReq:     commReq,
				ctx:         context.Background(),
				regInfo: createRegInfo("test", func(ctx context.Context, h *testproto.AccountHeader, r *testproto.DepositRequest) (*testproto.DepositResponse, error) {
					if r.Money != 100 {
						panic("req body not the value want pass")
					}
					if h.Account != "xiaoming" {
						panic("req header not the value want pass")
					}
					return nil, code.New(-1, "error")
				}),
			},
			wantErr: true,
		},

		{
			name: "3 input 1 output rsp error",
			args: args{
				bodyValue:   bodyValue,
				headerValue: headerValue,
				commReq:     commReq,
				ctx:         context.Background(),
				regInfo: createRegInfo("test", func(ctx context.Context, h *testproto.AccountHeader, r *testproto.DepositRequest) error {
					if r.Money != 100 {
						panic("req body not the value want pass")
					}
					if h.Account != "xiaoming" {
						panic("req header not the value want pass")
					}
					return code.NewMcode("ERROR", "error")
				}),
			},
			wantErr: true,
		},

		{
			name: "3 input 1 output rsp nil",
			args: args{
				bodyValue:   bodyValue,
				headerValue: headerValue,
				commReq:     commReq,
				ctx:         context.Background(),
				regInfo: createRegInfo("test", func(ctx context.Context, h *testproto.AccountHeader, r *testproto.DepositRequest) error {
					if r.Money != 100 {
						panic("req body not the value want pass")
					}
					if h.Account != "xiaoming" {
						panic("req header not the value want pass")
					}

					return nil
				}),
			},
			wantErr: false,
		},

		{
			name: "3 input 0 output",
			args: args{
				bodyValue:   bodyValue,
				headerValue: headerValue,
				commReq:     commReq,
				ctx:         context.Background(),
				regInfo:     createRegInfo("test", func(ctx context.Context, h *testproto.AccountHeader, r *testproto.DepositRequest) {}),
			},
			wantErr: false,
		},
		{
			name: "input commReq output commRsp",
			args: args{
				bodyValue:   bodyValue,
				headerValue: headerValue,
				commReq:     commReq,
				ctx:         context.Background(),
				regInfo: createRegInfo("test", func(ctx context.Context, req *grpccomm.CommRequest) *grpccomm.CommResponse {
					return &grpccomm.CommResponse{
						Result: true,
					}
				}),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp := call0(tt.args.ctx, tt.args.headerValue, tt.args.bodyValue, tt.args.commReq, tt.args.regInfo)
			if tt.wantErr == rsp.Result {
				t.Errorf("want err is %v,go result %v", tt.wantErr, rsp.Result)
				return
			}
		})
	}
}

func TestCheckSnowSlide(t *testing.T) {
	type args struct {
		showCount int32
	}
	tests := []struct {
		name    string
		args    args
		prepare func()
		wantErr bool
	}{

		{
			name: "snow protect re-counting",
			args: args{
				showCount: 10,
			},
			prepare: func() {
				// 计数时间低于当前时间，将会重置计数
				gCurTime = time.Now().Unix() - 1
			},
			wantErr: false,
		},
		{
			name: "snow protect same second",
			args: args{
				showCount: 10,
			},
			prepare: func() {
				// 保证本次检测的gCurTime和当前时间处于同一秒
				if (time.Now().Nanosecond() / int(time.Millisecond)) > 900 {
					time.Sleep(time.Millisecond * 100)
				}

				gCurTime = time.Now().Unix()
				gCurCount = 9
			},
			wantErr: false,
		},
		{
			name: "snow protect triggered",
			args: args{
				showCount: 10,
			},
			prepare: func() {
				gCurTime = time.Now().Unix() + 1
				gCurCount = 11
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()

			if err := checkSnowSlide(tt.args.showCount); (err != nil) != tt.wantErr {
				t.Errorf("checkSnowSlide() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
