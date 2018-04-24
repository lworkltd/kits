package consul

import (
	"fmt"
	"testing"
)

func Test_checkAndDefaultOption(t *testing.T) {
	type args struct {
		option *RegisterOption
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// succ
		{
			name: "http should be succ",
			args: args{
				option: &RegisterOption{
					ServerType:    ServerTypeHttp,
					Name:          "TestService",
					Id:            "TestService-master",
					Ip:            "10.0.0.1",
					Port:          8080,
					CheckInterval: "5s",
					CheckTimeout:  "15s",
				},
			},
		},
		{
			name: "grpc should be succ",
			args: args{
				option: &RegisterOption{
					ServerType:    ServerTypeGrpc,
					Name:          "TestService",
					Id:            "TestService-master",
					Ip:            "10.0.0.1",
					Port:          8080,
					CheckInterval: "5s",
					CheckTimeout:  "15s",
				},
			},
		},
		// failed
		{
			name: "should be missing the id",
			args: args{
				option: &RegisterOption{
					ServerType:    ServerTypeGrpc,
					Name:          "TestService",
					Id:            "",
					Ip:            "10.0.0.1",
					Port:          8080,
					CheckInterval: "5s",
					CheckTimeout:  "15s",
				},
			},
			wantErr: true,
		},
		{
			name: "should be missing the name",
			args: args{
				option: &RegisterOption{
					ServerType:    ServerTypeGrpc,
					Name:          "",
					Id:            "TestService-master",
					Ip:            "10.0.0.1",
					Port:          8080,
					CheckInterval: "5s",
					CheckTimeout:  "15s",
				},
			},
			wantErr: true,
		},
		{
			name: "should be missing the ip",
			args: args{
				option: &RegisterOption{
					ServerType:    ServerTypeGrpc,
					Name:          "TestService",
					Id:            "TestService-master",
					Ip:            "",
					Port:          8080,
					CheckInterval: "5s",
					CheckTimeout:  "15s",
				},
			},
			wantErr: true,
		},
		{
			name: "should missing the ip",
			args: args{
				option: &RegisterOption{
					ServerType:    ServerTypeGrpc,
					Name:          "TestService",
					Id:            "TestService-master",
					Ip:            "10.0.0.1",
					Port:          0,
					CheckInterval: "5s",
					CheckTimeout:  "15s",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkAndDefaultOption(tt.args.option); (err != nil) != tt.wantErr {
				t.Errorf("checkAndDefaultOption() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServerTypeString(t *testing.T) {
	tests := []struct {
		name       string
		serverType ServerType
		want       string
	}{
		{
			name: "http",
			serverType: func() ServerType {
				return ServerTypeHttp
			}(),
			want: "HttpServer",
		},
		{
			name: "grpc",
			serverType: func() ServerType {
				return ServerTypeGrpc
			}(),
			want: "GrpcServer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmt.Sprintf("%v", tt.serverType); got != tt.want {
				t.Errorf("ServerType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
