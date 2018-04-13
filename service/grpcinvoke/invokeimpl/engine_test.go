package invokeimpl

import (
	"testing"

	"github.com/lworkltd/kits/service/grpcinvoke"
)

func TestEngineNewGrpcAddr(t *testing.T) {
	type args struct {
		addr              string
		freeConnAfterUsed bool
	}
	tests := []struct {
		name   string
		engine *engine
		args   args
	}{
		{
			engine: newEngine(),
			args: args{
				addr:              "127.0.0.1",
				freeConnAfterUsed: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.engine.newAddr(tt.args.addr, tt.args.freeConnAfterUsed)
			got.Grpc("NoExistent-GRPC")
		})
	}
}

func TestEngineNewGrpcService(t *testing.T) {
	type args struct {
		serviceName       string
		discovery         grpcinvoke.DiscoveryFunc
		freeConnAfterUsed bool
	}
	tests := []struct {
		name   string
		engine *engine
		args   args
	}{
		{
			engine: newEngine(),
			args: args{
				serviceName:       "testService",
				discovery:         createAddrDiscovery("127.0.0.1:8080"),
				freeConnAfterUsed: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.engine.newService(tt.args.serviceName, tt.args.discovery, tt.args.freeConnAfterUsed)
			got.Grpc("No-Existential-GRPC")
		})
	}
}

func TestEngineNewAddr(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name   string
		engine *engine
		args   args
	}{
		{
			engine: newEngine(),
			args: args{
				addr: "127.0.0.1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.engine.Addr(tt.args.addr)
		})
	}
}

func TestEngineGrpcAddr(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name   string
		engine *engine
		args   args
	}{
		{
			engine: newEngine(),
			args: args{
				addr: "127.0.0.1:8080",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.engine.Addr(tt.args.addr); got == nil {
				t.Errorf("engine.GrpcAddr() return nil")
				return
			}
		})
	}
}

func TestEngineGrpcService(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		engine *engine
		args   args
	}{
		{
			engine: newEngine(),
			args: args{
				name: "test-service",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.engine.Service(tt.args.name); got == nil {
				t.Errorf("engine.GrpcService() return nil")
				return
			}
		})
	}
}
