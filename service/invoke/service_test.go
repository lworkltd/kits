package invoke

import (
	"fmt"
	"testing"
)

func Testservice_remote(t *testing.T) {
	svc := &service{
		discover: func(string) ([]string, error) {
			return []string{"127.0.0.1", "127.0.0.2"}, nil
		},
	}
	tests := []struct {
		name    string
		service *service
		want    string
		wantErr bool
	}{
		{
			name:    "firsttime",
			service: svc,
			want:    "127.0.0.1",
		},
		{
			name:    "secondtime",
			service: svc,
			want:    "127.0.0.2",
		},
		{
			name:    "thirdtime",
			service: svc,
			want:    "127.0.0.1",
		},
		{
			name: "emtpy",
			service: &service{
				discover: func(string) ([]string, error) {
					return []string{}, nil
				},
			},
			wantErr: true,
		},
		{
			name: "emtpy",
			service: &service{
				discover: func(string) ([]string, error) {
					return []string{"127.0.0.1:8080"}, nil
				},
			},
			want: "127.0.0.1:8080",
		},
		{
			name: "error",
			service: &service{
				discover: func(string) ([]string, error) {
					return []string{}, fmt.Errorf("service error")
				},
			},
			wantErr: true,
		},
		{
			name:    "nil_disconver",
			service: &service{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.service.remote()
			if (err != nil) != tt.wantErr {
				t.Errorf("service.remote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("service.remote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newRest(t *testing.T) {
	service := &service{
		discover: func(string) ([]string, error) {
			return []string{"127.0.0.1", "127.0.0.2"}, nil
		},
		name:       "auth_service",
		useTracing: true,
		useCircuit: true,
	}
	type args struct {
		service Service
		method  string
		path    string
		remote  string
		err     error
	}
	tests := []struct {
		name string
		args args
		want Client
	}{
		{
			args: args{
				service: service,
				method:  "PUT",
				path:    "/v1/country/{country}/province/{province}",
				remote:  "10.0.0.1:12034",
				err:     fmt.Errorf("remote not found"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newRest(tt.args.service, tt.args.method, tt.args.path, tt.args.remote, tt.args.err); got == nil {
				t.Errorf("newRest() = nil")
			}
		})
	}
}

func Testservice_Method(t *testing.T) {
	type args struct {
		method string
		path   string
	}
	tests := []struct {
		name    string
		service *service
		args    args
		wantNil bool
	}{
		{
			name: "error",
			service: &service{
				discover: func(string) ([]string, error) {
					return []string{"127.0.0.1:12304"}, nil
				},
			},
			wantNil: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.service.Method(tt.args.method, tt.args.path); (got == nil) != tt.wantNil {
				t.Errorf("service.Method() = %v, want %v", got, tt.wantNil)
			}
		})
	}
}

func Testservice_Get(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		service *service
		args    args
		wantNil bool
	}{
		{
			name: "error",
			service: &service{
				discover: func(string) ([]string, error) {
					return []string{"127.0.0.1:12304"}, nil
				},
			},
			wantNil: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.service.Get(tt.args.path); (got == nil) != tt.wantNil {
				t.Errorf("service.Get() = %v, want %v", got, tt.wantNil)
			}
		})
	}
}
