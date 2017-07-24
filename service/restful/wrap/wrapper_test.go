package wrap

import (
	"reflect"
	"testing"

	"github.com/lvhuat/kits/service/restful/code"
)

func TestWrapper_Done(t *testing.T) {
	type args struct {
		v []interface{}
	}
	tests := []struct {
		name    string
		wrapper *Wrapper
		args    args
		result  bool
		message string
		Mcode   string
		Data    interface{}
	}{
		{
			wrapper: NewWrapper("MYSERVICE_"),
			result:  true,
			message: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.Done(tt.args.v...); !reflect.DeepEqual(got.Result(), tt.result) {
				t.Errorf("Wrapper.Done() = %v, want %v", got.Result(), tt.result)
			}
		})
	}
}

func TestWrapper_FromError(t *testing.T) {
	type args struct {
		cerr code.Error
	}
	tests := []struct {
		name    string
		wrapper *Wrapper
		args    args
		result  bool
		message string
		mcode   string
	}{
		{
			wrapper: NewWrapper("MYSERVICE_"),
			args: args{
				cerr: code.New("XXX_XXX", "local"),
			},
			result:  false,
			message: "local",
			mcode:   "XXX_XXX",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.FromError(tt.args.cerr); !reflect.DeepEqual(got.Result(), tt.result) {
				t.Errorf("Wrapper.FromError() result = %v, want %v", got.Result(), tt.result)
			}
			if got := tt.wrapper.FromError(tt.args.cerr); !reflect.DeepEqual(got.Message(), tt.message) {
				t.Errorf("Wrapper.FromError() message = %v, want %v", got.Message(), tt.message)
			}
			if got := tt.wrapper.FromError(tt.args.cerr); !reflect.DeepEqual(got.Mcode(), tt.mcode) {
				t.Errorf("Wrapper.FromError() mcode = %v, want %v", got.Mcode(), tt.mcode)
			}
		})
	}
}
