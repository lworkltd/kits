package svc

import "testing"

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
		{},
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
