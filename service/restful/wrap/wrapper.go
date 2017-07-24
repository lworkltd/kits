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

// NewWrapper 创建一个新的wrapper
func NewWrapper(mcodePrefix string) *Wrapper {
	return &Wrapper{
		mcodePrefix: mcodePrefix,
		pool: sync.Pool{
			New: func() interface{} {
				return new(WrappedResponse)
			},
		},
	}
}

// WrappedFunc 是用于封装GIN HTTP接口返回为通用接口的函数定义
type WrappedFunc func(ctx *gin.Context) Response

// Wrap 为gin的回调接口增加了固定的返回值，当程序收到处理结果的时候会将返回值封装一层再发送到网络
func (wrapper *Wrapper) Wrap(f WrappedFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		r := f(ctx)

		var wrappedResp interface{}
		if r.Result() != true {
			wrappedResp = map[string]interface{}{
				"mcode":  r.Mcode(),
				"result": r.Result(),
				"msg":    r.Message(),
			}
		}

		wrappedResp = map[string]interface{}{
			"result": r.Result(),
			"data":   r.Data(),
		}

		wrapper.pool.Put(r)

		ctx.JSON(200, wrappedResp)
	}
}

func (wrapper *Wrapper) Handle(method string, eng *gin.Engine, path string, f WrappedFunc) {
	eng.Handle(method, path, wrapper.Wrap(f))
}

func (wrapper *Wrapper) Get(eng *gin.Engine, path string, f WrappedFunc) {
	wrapper.Handle("GET", eng, path, f)
}

func (wrapper *Wrapper) Patch(eng *gin.Engine, path string, f WrappedFunc) {
	wrapper.Handle("POST", eng, path, f)
}

func (wrapper *Wrapper) Post(eng *gin.Engine, path string, f WrappedFunc) {
	wrapper.Handle("DELETE", eng, path, f)
}

func (wrapper *Wrapper) Put(eng *gin.Engine, path string, f WrappedFunc) {
	wrapper.Handle("PUT", eng, path, f)
}

func (wrapper *Wrapper) Options(eng *gin.Engine, path string, f WrappedFunc) {
	wrapper.Handle("OPTIONS", eng, path, f)
}

func (wrapper *Wrapper) Head(eng *gin.Engine, path string, f WrappedFunc) {
	wrapper.Handle("HEAD", eng, path, f)
}

// Error 失败并且打印指定实例
func (wrapper *Wrapper) Error(mcode int, m interface{}) Response {
	r := wrapper.pool.Get().(*WrappedResponse)
	r.result = false
	r.message = fmt.Sprint(m)
	r.mcode = fmt.Sprintf("%s_%d", wrapper.mcodePrefix, mcode)

	return r
}

// Errorln 失败并且打印指定实例列表
func (wrapper *Wrapper) Errorln(mcode int, ms ...interface{}) Response {
	r := wrapper.pool.Get().(*WrappedResponse)
	r.result = false
	r.message = fmt.Sprint(ms...)
	r.mcode = fmt.Sprintf("%s_%d", wrapper.mcodePrefix, mcode)

	return r
}

// Errorf 失败并且按格式打印指定内容
func (wrapper *Wrapper) Errorf(mcode int, format string, args ...interface{}) Response {
	r := wrapper.pool.Get().(*WrappedResponse)
	r.result = false
	r.message = fmt.Sprintf(format, args...)
	r.mcode = fmt.Sprintf("%s_%d", wrapper.mcodePrefix, mcode)

	return r
}

// FromError 通过Error创建一个实例
func (wrapper *Wrapper) FromError(cerr code.Error) Response {
	r := wrapper.pool.Get().(*WrappedResponse)
	r.result = (cerr == nil)
	r.message = cerr.Message()
	r.mcode = cerr.Code()

	return r
}

// Done 成功并且返回数据
// 如果什么都不传，则不返回数据
// 如果传递数据，则取第一个数据作为data
// 因此，此处并不建议使用多个值作为参数，如果你想返回一个数组，那么你就直接把数据作为第一个参数
func (wrapper *Wrapper) Done(v ...interface{}) Response {
	var data interface{}
	if len(v) > 1 {
		panic("bad parameter length, please pass parameter less than one")
	}

	if len(v) != 0 {
		data = v[0]
	}

	r := wrapper.pool.Get().(*WrappedResponse)

	r.result = true
	r.data = data

	return r
}
