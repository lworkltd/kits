package wrap

import (
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
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
			wrapper: New("MYSERVICE_"),
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
			wrapper: New("MYSERVICE_"),
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

func Test_Server(t *testing.T) {
	FailedCode := 10010

	type Data struct {
		Name string
		Age  int
	}

	wrapper := New("MYSERVICE_EXCEPTION_")
	foo := func(c *gin.Context) Response {
		routeError := c.Params.ByName("error")
		if routeError == "yes" {
			return wrapper.Errorf(FailedCode, "Foo failed! %v", routeError)
		}

		ret := &Data{
			Name: "Anna",
			Age:  15,
		}
		return wrapper.Done(ret)
	}

	bar := func(c *gin.Context) Response {
		routeError := c.Params.ByName("error")
		if routeError == "yes" {
			return wrapper.Error(FailedCode, "Bar Failed")
		}

		return wrapper.Done(&Data{
			Name: "Petter",
			Age:  32,
		})
	}

	r := gin.Default()
	wrapper.Get(r, "/foo", foo)

	v2 := r.Group("/v2")
	wrapper.Post(v2, "/bar", bar)
	wrapper.Get(v2, "/bar", bar)
	wrapper.Put(v2, "/bar", bar)
	wrapper.Options(v2, "/bar", bar)
	wrapper.Patch(v2, "/bar", bar)
	wrapper.Head(v2, "/bar", bar)
}
