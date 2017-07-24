package wrap

import (
	"reflect"
	"testing"
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
