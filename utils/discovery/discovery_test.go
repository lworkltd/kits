package discovery

import "testing"

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
