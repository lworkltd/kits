package invoke

import (
	"fmt"
	"testing"
)

func TestServiceRemote(t *testing.T) {
	svc := &service{
		discovery: func(string) ([]string, []string, error) {
			return []string{"127.0.0.1", "127.0.0.2"}, []string{"my-service-1", "my-service-2"}, nil
		},
	}

	svc1 := &service{
		discovery: func(string) ([]string, []string, error) {
			return []string{"127.0.0.1"}, []string{"my-service-1"}, nil
		},
	}
	tests := []struct {
		name    string
		service *service
		want    string
		want1   string
		wantErr bool
	}{
		{
			name:    "firsttime",
			service: svc1,
			want:    "127.0.0.1",
			want1:   "my-service-1",
		},
		{
			name:    "firsttime",
			service: svc,
			want:    "127.0.0.1",
			want1:   "my-service-1",
		},
		{
			name:    "secondtime",
			service: svc,
			want:    "127.0.0.2",
			want1:   "my-service-2",
		},
		{
			name:    "thirdtime",
			service: svc,
			want:    "127.0.0.1",
			want1:   "my-service-1",
		},
		{
			name: "emtpy",
			service: &service{
				discovery: func(string) ([]string, []string, error) {
					return []string{}, []string{}, nil
				},
			},
			wantErr: true,
		},
		{
			name: "single",
			service: &service{
				discovery: func(string) ([]string, []string, error) {
					return []string{"127.0.0.1:8080"}, []string{"service"}, nil
				},
			},
			want:  "127.0.0.1:8080",
			want1: "service",
		},
		{
			name: "error",
			service: &service{
				discovery: func(string) ([]string, []string, error) {
					return []string{}, []string{}, fmt.Errorf("service error")
				},
			},
			wantErr: true,
		},
		{
			name:    "nil_discover",
			service: &service{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.service.remote()
			if (err != nil) != tt.wantErr {
				t.Errorf("service.remote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("service.remote() = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("service.remote() = %v, want1 %v", got, tt.want1)
			}
		})
	}
}

func TestNewRest(t *testing.T) {
	service := &service{
		discovery: func(string) ([]string, []string, error) {
			return []string{"127.0.0.1", "127.0.0.2"}, []string{"my-service-1", "my-service-2"}, nil
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
		id      string
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
			if got := newRest(tt.args.service, tt.args.method, tt.args.path, tt.args.remote, tt.args.id, tt.args.err); got == nil {
				t.Errorf("newRest() = nil")
			}
		})
	}
}

func TestServiceMethod(t *testing.T) {
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
				discovery: func(string) ([]string, []string, error) {
					return []string{"127.0.0.1:12304"}, []string{"service-id"}, nil
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

func TestServiceGet(t *testing.T) {
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
				discovery: func(string) ([]string, []string, error) {
					return []string{"127.0.0.1:12304"}, []string{"service-id"}, nil
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
