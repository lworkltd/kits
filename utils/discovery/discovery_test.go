package discovery

import (
	"testing"

	"github.com/lworkltd/kits/helper/consul"
	"github.com/lworkltd/kits/service/discovery"
	"github.com/lworkltd/kits/service/profile"
)

func init() {
	discovery.Init(&discovery.Option{
		SearchFunc:   func(string) ([]string, []string, error) { return nil, nil, nil },
		RegisterFunc: func(*consul.RegisterOption) error { return nil },
	})
}
func TestMakeCheckUrl(t *testing.T) {
	type args struct {
		ip   string
		port int
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "url",
			args: args{
				ip:   "127.0.0.1",
				port: 8080,
				path: "http://127.0.0.2:8080/health",
			},
			want: "http://127.0.0.2:8080/health",
		},

		{
			name: "path",
			args: args{
				ip:   "127.0.0.3",
				port: 8080,
				path: "/health",
			},
			want: "http://127.0.0.3:8080/health",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeCheckUrl(tt.args.ip, tt.args.port, tt.args.path); got != tt.want {
				t.Errorf("makeCheckUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckAndResolveProfile(t *testing.T) {
	type args struct {
		cfg *profile.Service
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				cfg: &profile.Service{
					Host:       ":8080",
					ReportIp:   "192.168.0.1",
					ReportName: "my-service",
					ReportId:   "my-service-1",
				},
			},
			want:    8080,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkAndResolveProfile(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkAndResolveProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkAndResolveProfile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegisterServerWithProfile(t *testing.T) {
	type args struct {
		checkUrl string
		cfg      *profile.Service
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				cfg: &profile.Service{
					Reportable: true,
					Host:       ":8080",
					ReportIp:   "192.168.0.1",
					ReportName: "my-service",
					ReportId:   "my-service-1",
				},
				checkUrl: "/health",
			},
			wantErr: false,
		},
		{
			name: "no-register",
			args: args{
				cfg: &profile.Service{
					Reportable: false,
				},
			},
			wantErr: false,
		},
		{
			name: "lack-port",
			args: args{
				cfg: &profile.Service{
					Reportable: true,
					Host:       "",
					ReportIp:   "192.168.0.1",
					ReportName: "my-service",
					ReportId:   "my-service-1",
				},
				checkUrl: "/health",
			},
			wantErr: true,
		},
		{
			name: "lack-ip",
			args: args{
				cfg: &profile.Service{
					Reportable: true,
					Host:       ":8080",
					ReportIp:   "",
					ReportName: "my-service",
					ReportId:   "my-service-1",
				},
				checkUrl: "/health",
			},
			wantErr: true,
		},
		{
			name: "lack-service",
			args: args{
				cfg: &profile.Service{
					Reportable: true,
					Host:       ":8080",
					ReportIp:   "192.168.0.1",
					ReportName: "",
					ReportId:   "my-service-1",
				},
				checkUrl: "/health",
			},
			wantErr: true,
		},
		{
			name: "lack-service-id",
			args: args{
				cfg: &profile.Service{
					Reportable: true,
					Host:       ":8080",
					ReportIp:   "192.168.0.1",
					ReportName: "my-service",
					ReportId:   "",
				},
				checkUrl: "/health",
			},
			wantErr: true,
		},
		{
			name: "lack-check-url",
			args: args{
				cfg: &profile.Service{
					Reportable: true,
					Host:       ":8080",
					ReportIp:   "192.168.0.1",
					ReportName: "my-service",
					ReportId:   "",
				},
				checkUrl: "",
			},
			wantErr: true,
		},
		{
			name: "port-no-number",
			args: args{
				cfg: &profile.Service{
					Reportable: true,
					Host:       ":abc",
					ReportIp:   "192.168.0.1",
					ReportName: "my-service",
					ReportId:   "",
				},
				checkUrl: "",
			},
			wantErr: true,
		},
		{
			name: "loopback-forbiden",
			args: args{
				cfg: &profile.Service{
					Reportable: true,
					Host:       ":abc",
					ReportIp:   "localhost",
					ReportName: "my-service",
					ReportId:   "",
				},
				checkUrl: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RegisterServerWithProfile(tt.args.checkUrl, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("RegisterServerWithProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
