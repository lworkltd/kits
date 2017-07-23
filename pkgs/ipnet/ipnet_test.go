package ipnet

import "testing"

func TestIp(t *testing.T) {
	type args struct {
		adapter []string
	}
	tests := []struct {
		name    string
		args    args
		want1   bool
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				adapter: []string{},
			},
			want1:   true,
			wantErr: false,
		},
		{
			name: "not_found",
			args: args{
				adapter: []string{"not_found_interface"},
			},
			want1:   false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := Ipv4(tt.args.adapter...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got1 && len(got) <= 0 {
				t.Errorf("Ip() got = %v", got)
			}
			if got1 != tt.want1 {
				t.Errorf("Ip() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
