package httpsrv

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	// DefaultSnowSlideLimit  默认过载保护
	DefaultSnowSlideLimit int32 = 20000
)

// Option 服务的配置选项
type Option struct {
	WrapFunc       func(f interface{}) gin.HandlerFunc // 自定义函数转换，将自定义的处理函数转化成为gin框架的处理函数，如果设置了这个选项，其他选项都不会生效
	Prefix         string                              // 如果返回的错误码中只有code，就会使用prefix创建一个mcode，格式：<Prefix>_<code>
	Logger         *logrus.Logger                      // 默认日志打印对象，不传就会使用logrus.StdLogger
	SnowSlide      SnowSlide                           // 自定义防雪崩拦截器，如果不设置就会使用SnowSlideLimit创建一个默认的
	SnowSlideLimit int32                               // 使用默认的SnowSlide，表示最大未返回的请求数，SnowSlide未设置时有效
	Report         ReportFunc                          // 自定义上报函数
}

func (option *Option) useDefault() {
	if option.SnowSlide == nil {
		if option.SnowSlideLimit <= 0 {
			option.SnowSlide = &snowSlide{
				LimitCnt: DefaultSnowSlideLimit,
				Service:  option.Prefix,
			}
		} else {
			option.SnowSlide = &snowSlide{
				LimitCnt: option.SnowSlideLimit,
				Service:  option.Prefix,
			}
		}
	}

	if option.Report == nil {
		option.Report = defaultReportProcessResultToMonitor
	}

	if option.Logger == nil {
		option.Logger = logrus.StandardLogger()
	}
}

// Wrapper 用于对请求返回结果进行封装的类
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
	// 日志打印对象
	logger *logrus.Logger
	// 请求熔断拦截
	snowSlide SnowSlide
	// 处理函数转换
	wrapFunc func(f interface{}) gin.HandlerFunc
	// 上报处理
	report ReportFunc
}

// New 创建一个新的wrapper
func New(option *Option) *Wrapper {
	option.useDefault()

	w := &Wrapper{
		mcodePrefix: option.Prefix,
		pool: sync.Pool{
			New: func() interface{} {
				return new(Response)
			},
		},
		snowSlide: option.SnowSlide,
		wrapFunc:  option.WrapFunc,
		report:    option.Report,
		logger:    option.Logger,
	}
	if w.wrapFunc == nil {
		w.wrapFunc = w.defaultWrap
	}

	return w
}

// WrapFunc 封装函数返回一个能够被gin直接使用的回调函数,提供足够的伸缩新
// f 是自定义的回调函数
// path 是HTTP的路径
// 以DefaultWrappedFunc作为f的类型为例：
// func(f interface{}) gin.HandlerFunc {
//	return func(ginCtx *gin.Context){
//		realf := f.(func(ginCtx*gin.Context)(interface,error)
//		retData,err:= realf(ginCtx)
// 		...
// 	}
// }
type WrapFunc func(f interface{}) gin.HandlerFunc

// HttpServer 抽象gin的Group和Root
type HttpServer interface {
	Handle(string, string, ...gin.HandlerFunc) gin.IRoutes
}

// Handle Http-通用请求注册
func (wrapper *Wrapper) Handle(method string, srv HttpServer, path string, f interface{}) {
	srv.Handle(method, path, wrapper.wrapFunc(f))
}

// Get Http-GET请求注册
func (wrapper *Wrapper) Get(srv HttpServer, path string, f interface{}) {
	wrapper.Handle("GET", srv, path, f)
}

// Patch Http-PATCH请求注册
func (wrapper *Wrapper) Patch(srv HttpServer, path string, f interface{}) {
	wrapper.Handle("PATCH", srv, path, f)
}

// Post Http-POST请求注册
func (wrapper *Wrapper) Post(srv HttpServer, path string, f interface{}) {
	wrapper.Handle("POST", srv, path, f)
}

// Put Http-PUT请求注册
func (wrapper *Wrapper) Put(srv HttpServer, path string, f interface{}) {
	wrapper.Handle("PUT", srv, path, f)
}

// Options Http-Options请求注册
func (wrapper *Wrapper) Options(srv HttpServer, path string, f interface{}) {
	wrapper.Handle("OPTIONS", srv, path, f)
}

// Head Http-HEADER请求注册
func (wrapper *Wrapper) Head(srv HttpServer, path string, f interface{}) {
	wrapper.Handle("HEAD", srv, path, f)
}

// Delete Http-DELETE请求注册
func (wrapper *Wrapper) Delete(srv HttpServer, path string, f interface{}) {
	wrapper.Handle("DELETE", srv, path, f)
}

// Response 默认的返回
type Response struct {
	Result    bool        `json:"result"`
	Mcode     string      `json:"mcode,omitempty"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
}
