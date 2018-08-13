package urpccomm

import (
	"testing"
)

func TestParseCodeError(t *testing.T) {
	type args struct {
		mcode string
		msg   string
	}
	tests := []struct {
		name       string
		args       args
		wantCode   int
		wantError  string
		wantMcode  string
		wantPrefix string
		wantMsg    string
	}{
		{
			name: "not int error",
			args: args{
				mcode: "NOT_INT_ERROR",
				msg:   "not int error",
			},
			wantCode:   NotIntCodeError,
			wantError:  "NOT_INT_ERROR,not int error",
			wantMcode:  "NOT_INT_ERROR",
			wantPrefix: "NOT_INT_ERROR",
			wantMsg:    "not int error",
		},
		{
			name: "int error",
			args: args{
				mcode: "MYSERVICE_1003",
				msg:   "int error",
			},
			wantCode:   1003,
			wantError:  "MYSERVICE_1003,int error",
			wantMcode:  "MYSERVICE_1003",
			wantPrefix: "MYSERVICE",
			wantMsg:    "int error",
		},

		{
			name: "tail with underline",
			args: args{
				mcode: "MYSERVICE_",
				msg:   "tail with underline",
			},
			wantCode:   NotIntCodeError,
			wantError:  "MYSERVICE_,tail with underline",
			wantMcode:  "MYSERVICE_",
			wantPrefix: "MYSERVICE_",
			wantMsg:    "tail with underline",
		},
		{
			name: "mcode empty",
			args: args{
				mcode: "",
				msg:   "mcode empty",
			},
			wantCode:   NotIntCodeError,
			wantError:  "UNKOWN_ERROR,mcode empty",
			wantMcode:  "UNKOWN_ERROR",
			wantPrefix: "",
			wantMsg:    "mcode empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCodeError(tt.args.mcode, tt.args.msg)
			if tt.wantCode != got.Code() {
				t.Errorf("want Code = %d,got %d", tt.wantCode, got.Code())
				return
			}

			if tt.wantError != got.Error() {
				t.Errorf("want Error = %s,got %s", tt.wantCode, got.Error())
				return
			}

			if tt.wantMcode != got.Mcode() {
				t.Errorf("want Mcode = %d,got %d", tt.wantCode, got.Code())
				return
			}

			if tt.wantPrefix != got.prefix {
				t.Errorf("want prefix = %s,got %s", tt.wantCode, got.prefix)
				return
			}

			if tt.wantMsg != got.message {
				t.Errorf("want msg = %s,got %s", tt.wantCode, got.message)
				return
			}

		})
	}
}
