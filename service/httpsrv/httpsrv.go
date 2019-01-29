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
	// 如果返回的错误码中只有code，就会使用prefix创建一个mcode，格式：<Prefix>_<code>
	Prefix string

	// 以下为自定义函数，可以根据自己的需要修改，但是出现panic需要自己负责
	// 默认日志打印对象，不传就会使用logrus.StdLogger
	Logger *logrus.Logger
	// 自定义防雪崩拦截器，如果不设置就会使用SnowSlideLimit创建一个默认的
	SnowSlide SnowSlide
	// 使用默认的SnowSlide，表示最大未返回的请求数，SnowSlide未设置时有效
	SnowSlideLimit int32
	// 自定义上报函数
	Report ReportFunc
	// 打印日志函数
	LogFunc LogFunc
	// 将返回写到网络IO中
	WriteResult WriteResultFunc

	// 不建议使用，自定义转换函数，修改本函数后会重新定义本函数后，其他的选项均不生效，需要使用者自己全部重写
	// 使用者如果需要注册自定义的函数格式使使用
	WrapFunc func(f interface{}) gin.HandlerFunc
}

// useDefault 补全配置
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
		option.Report = DefaultReport
	}

	if option.Logger == nil {
		option.Logger = logrus.StandardLogger()
	}

	if option.LogFunc == nil {
		option.LogFunc = DefaultLogFunc
	}

	if option.WriteResult == nil {
		option.WriteResult = DefaultWriteResultFunc
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
	// 日志处理函数
	logFunc LogFunc
	// 结果IO处理
	writeResult WriteResultFunc
}

// New 创建一个新的wrapper
func New(option *Option) *Wrapper {
	option.useDefault()

	w := &Wrapper{
		mcodePrefix: option.Prefix,
		pool: sync.Pool{
			New: func() interface{} {
				return new(DefaultResponse)
			},
		},
		snowSlide:   option.SnowSlide,
		wrapFunc:    option.WrapFunc,
		report:      option.Report,
		logger:      option.Logger,
		writeResult: option.WriteResult,
		logFunc:     option.LogFunc,
	}

	if w.wrapFunc == nil {
		w.wrapFunc = w.DefaultWrap
	}

	return w
}

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

// DefaultResponse 默认的返回
type DefaultResponse struct {
	Result    bool        `json:"result"`
	Mcode     string      `json:"mcode,omitempty"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
}
