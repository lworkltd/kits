package invoke

import (
	"testing"
)

func Test_makeUrl(t *testing.T) {
	type args struct {
		sche   string
		host   string
		path   string
		querys map[string][]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				sche: "http",
				host: "127.0.0.1:8013",
				path: "/v1/apples/total_weight",
				querys: map[string][]string{
					"box":   {"1", "2"},
					"color": []string{"yellow"},
				},
			},
			want:    "http://127.0.0.1:8013/v1/apples/total_weight?box=1&box=2&color=yellow",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeUrl(tt.args.sche, tt.args.host, tt.args.path, tt.args.querys)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("makeUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parsePath(t *testing.T) {
	type args struct {
		path string
		r    map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				path: "/v1/apples/{apple}/total_weight",
				r: map[string]string{
					"apple": "1",
				},
			},
			want:    "/v1/apples/1/total_weight",
			wantErr: false,
		},
		{
			name: "normal",
			args: args{
				path: "/v1/apples/{apple}/cololrs/{colors}/state/{fresh_state}",
				r: map[string]string{
					"apple":       "1",
					"colors":      "red",
					"fresh_state": "best",
				},
			},
			want:    "/v1/apples/1/cololrs/red/state/best",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePath(tt.args.path, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parsePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
