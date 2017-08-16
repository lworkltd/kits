package wrap

import (
	"fmt"
	"log"
	"os"

	"sync"

	"net/http"
	"runtime/debug"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/context"
	"github.com/lworkltd/kits/service/restful/code"
	logutils "github.com/lworkltd/kits/utils/log"
)

// Wrapper 用于对请求返回结果进行封装的类
// TODO:需要增加单元测试 wrapper_test.go
type Wrapper struct {
	// 错误码的前缀
	// 比如 错误码为1001，前缀为ANYPROJECT_ANYSERVICE_,那么返回给调用者的错误码(mcode)就为:ANYPROJECT_ANYSERVICE_1001
	mcodePrefix string
	// 返回对象和回收，高并发场景下的内存重复利用 变[use->gc->allocate manager->use] 为 [use->pool->use]
	pool sync.Pool
	// 模式
	mode string
	// 服务名称
	serviceName string
	// 服务ID
	serviceId string
}

type Option struct {
	Prefix string
	Mode   string
}

// NewWrapper 创建一个新的wrapper
func New(option *Option) *Wrapper {
	w := &Wrapper{
		mcodePrefix: option.Prefix,
		pool: sync.Pool{
			New: func() interface{} {
				return new(Response)
			},
		},
	}

	return w
}

// WrappedFunc 是用于封装GIN HTTP接口返回为通用接口的函数定义
type WrappedFunc func(srvContext context.Context, ctx *gin.Context) (interface{}, code.Error)

type HttpServer interface {
	Handle(string, string, ...gin.HandlerFunc) gin.IRoutes
}

// Wrap 为gin的回调接口增加了固定的返回值，当程序收到处理结果的时候会将返回值封装一层再发送到网络
func (wrapper *Wrapper) Wrap(f WrappedFunc) gin.HandlerFunc {
	return func(httpCtx *gin.Context) {
		Prefix := wrapper.mcodePrefix // 错误码前缀
		logger := logrus.New()
		logger.Out = os.Stderr
		// 设置日志等级
		logger.Level = logrus.InfoLevel
		// 设置日志格式,让附加的TAG放在最前面
		formatter := &logutils.TextFormatter{
			TimestampFormat: "01-02 15:04:05.999",
		}
		logger.Formatter = formatter
		// 附加服务ID
		logger.Hooks.Add(logutils.NewServiceTagHook(wrapper.serviceName, wrapper.serviceId, wrapper.mode))
		// 附加日志文件行号
		logger.Hooks.Add(logutils.NewFileLineHook(log.Lshortfile))
		// 附加Tracing TAG
		logger.Hooks.Add(logutils.NewTracingLogHook())
		serviceCtx := context.FromHttpRequest(httpCtx.Request, logger)
		defer serviceCtx.Finish()

		// 附加Tracing Id
		tracingHeader := http.Header{}
		serviceCtx.Inject(tracingHeader)
		logger.Hooks.Add(logutils.NewTracingTagHook(serviceCtx.TracingId()))

		since := time.Now()
		var (
			data interface{}
			cerr code.Error
		)
		defer func() {
			// 拦截业务层的异常
			if r := recover(); r != nil {
				fmt.Println(r)
				if codeErr, ok := r.(code.Error); ok {
					cerr = codeErr
				} else {
					cerr = code.New(100000000, "Service internal error")
					serviceCtx.WithFields(logrus.Fields{
						"error": r,
						"stack": string(debug.Stack()),
					}).Errorln("Panic")
				}
			}
			// 错误的返回
			if cerr != nil {
				httpCtx.JSON(200, map[string]interface{}{
					"result":  false,
					"mcode":   fmt.Sprintf("%s_%d", Prefix, cerr.Code()),
					"message": cerr.Error(),
				})
			} else {
				httpCtx.JSON(200, map[string]interface{}{
					"result": true,
					"data":   data,
				})
			}
			// 正确的返回
			l := serviceCtx.WithFields(logrus.Fields{
				"method": httpCtx.Request.Method,
				"path":   httpCtx.Request.URL.Path,
				"delay":  time.Since(since),
			})

			if cerr != nil {
				l.WithFields(logrus.Fields{
					"mcode":   fmt.Sprintf("%s_%d", Prefix, cerr.Code()),
					"message": cerr.Error(),
				}).Error("Http request failed")
			} else {
				l.Info("HTTP request done")
			}
		}()
		data, cerr = f(serviceCtx, httpCtx)
	}
}

func (wrapper *Wrapper) Handle(method string, srv HttpServer, path string, f WrappedFunc) {
	srv.Handle(method, path, wrapper.Wrap(f))
}

func (wrapper *Wrapper) Get(srv HttpServer, path string, f WrappedFunc) {
	wrapper.Handle("GET", srv, path, f)
}

func (wrapper *Wrapper) Patch(srv HttpServer, path string, f WrappedFunc) {
	wrapper.Handle("PATCH", srv, path, f)
}

func (wrapper *Wrapper) Post(srv HttpServer, path string, f WrappedFunc) {
	wrapper.Handle("POST", srv, path, f)
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

func (wrapper *Wrapper) Delete(srv HttpServer, path string, f WrappedFunc) {
	wrapper.Handle("DELETE", srv, path, f)
}
