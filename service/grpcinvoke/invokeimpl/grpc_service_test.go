package invokeimpl

import (
	"testing"

	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
)

func TestGrpcServiceGrpc(t *testing.T) {
	type args struct {
		callName string
	}
	tests := []struct {
		name        string
		grpcService *grpcService
		args        args
	}{
		{
			grpcService: &grpcService{
				name:              "grpc-service",
				freeConnAfterUsed: false,
				connLb: newGrpcConnBalancer("grpc-service", 4, func(string) ([]string, []string, error) {
					return []string{}, []string{}, nil
				}),
			},
			args: args{"GRPC_CALL_REQUEST"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.grpcService.Grpc(tt.args.callName)
			got.Body(&grpccomm.InvokeTestEchoRequest{}).
				Header(&grpccomm.CommHeader{}).
				Fallback(func(error) error { return nil }).
				Response(nil)
		})
	}
}

func TestGrpcConnBalancerClose(t *testing.T) {
	tests := []struct {
		name             string
		grpcConnBalancer *GrpcConnBalancer
		wantErr          bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.grpcConnBalancer.Close(); (err != nil) != tt.wantErr {
				t.Errorf("GrpcConnBalancer.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGrpcServiceClose(t *testing.T) {
	tests := []struct {
		name        string
		grpcService *grpcService
	}{
		{
			grpcService: &grpcService{
				name:              "grpc-service",
				freeConnAfterUsed: false,
				connLb: newGrpcConnBalancer("grpc-service", 4, func(string) ([]string, []string, error) {
					return []string{}, []string{}, nil
				}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.grpcService.Close()
			tt.grpcService.Close()
		})
	}
}
