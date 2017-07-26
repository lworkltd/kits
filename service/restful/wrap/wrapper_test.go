package wrap

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/restful/code"
)

func Test_Server(t *testing.T) {
	FailedCode := 10010

	type Data struct {
		Name string
		Age  int
	}

	wrapper := New(&Option{
		Prefix: "MYSERVICE_EXCEPTION_",
	})
	foo := func(c *gin.Context) (interface{}, code.Error) {
		routeError := c.Params.ByName("error")
		if routeError == "yes" {
			return nil, code.New(FailedCode, "Foo failed!")
		}

		ret := &Data{
			Name: "Anna",
			Age:  15,
		}
		return ret, nil
	}

	bar := func(c *gin.Context) (interface{}, code.Error) {
		routeError := c.Params.ByName("error")
		if routeError == "yes" {
			return nil, code.New(FailedCode, "Bar failed!")
		}

		return &Data{
			Name: "Petter",
			Age:  32,
		}, nil
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
	wrapper.Delete(v2, "/bar", bar)
}
