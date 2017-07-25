package wrap

import (
	"fmt"

	"sync"

	"github.com/gin-gonic/gin"
	"github.com/lvhuat/kits/service/restful/code"
)

// Wrapper 用于对请求返回结果进行封装的类
// TODO:需要增加单元测试 wrapper_test.go
type Wrapper struct {
	// 错误码的前缀
	// 比如 错误码为1001，前缀为ANYPROJECT_ANYSERVICE_,那么返回给调用者的错误码(mcode)就为:ANYPROJECT_ANYSERVICE_1001
	mcodePrefix string
	// 返回对象和回收，高并发场景下的内存重复利用 变[use->gc->allocate manager->use] 为 [use->pool->use]
	pool sync.Pool
}

type Option struct {
	Prefix string
}

// NewWrapper 创建一个新的wrapper
func New(option *Option) *Wrapper {
	return &Wrapper{
		mcodePrefix: option.Prefix,
		pool: sync.Pool{
			New: func() interface{} {
				return new(Response)
			},
		},
	}
}

// WrappedFunc 是用于封装GIN HTTP接口返回为通用接口的函数定义
type WrappedFunc func(ctx *gin.Context) (interface{}, code.Error)

type HttpServer interface {
	Handle(string, string, ...gin.HandlerFunc) gin.IRoutes
}

// Wrap 为gin的回调接口增加了固定的返回值，当程序收到处理结果的时候会将返回值封装一层再发送到网络
func (wrapper *Wrapper) Wrap(f WrappedFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		r, cerr := f(ctx)

		res := &Response{
			Data:   r,
			Result: true,
		}

		if cerr != nil {
			res.Result = false
			if cerr.Mcode() != "" {
				res.Mcode = cerr.Mcode()
			} else {
				res.Mcode = fmt.Sprintf("%s_%d", wrapper.mcodePrefix, cerr.Code())
			}
		}

		ctx.JSON(200, res)
	}
}

func (wrapper *Wrapper) Handle(method string, srv HttpServer, path string, f WrappedFunc) {
	srv.Handle(method, path, wrapper.Wrap(f))
}

func (wrapper *Wrapper) Get(srv HttpServer, path string, f WrappedFunc) {
	wrapper.Handle("GET", srv, path, f)
}

func (wrapper *Wrapper) Patch(srv HttpServer, path string, f WrappedFunc) {
	wrapper.Handle("POST", srv, path, f)
}

func (wrapper *Wrapper) Post(srv HttpServer, path string, f WrappedFunc) {
	wrapper.Handle("DELETE", srv, path, f)
}

func (wrapper *Wrapper) Put(srv HttpServer, path string, f WrappedFunc) {
	wrapper.Handle("PUT", srv, path, f)
}

func (wrapper *Wrapper) Options(srv HttpServer, path string, f WrappedFunc) {
	wrapper.Handle("OPTIONS", srv, path, f)
}

func (wrapper *Wrapper) Head(srv HttpServer, path string, f WrappedFunc) {
	wrapper.Handle("HEAD", srv, path, f)
}
