package httpsrv

import (
	"net/http"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/httpsrv/httpstat"
	"github.com/sirupsen/logrus"
)

var (
	// DefaultSnowSlideLimit  默认过载保护
	DefaultSnowSlideLimit int32 = 20000
)

// Option 服务的配置选项
type Option struct {
	// 如果返回的错误码中只有code，就会使用prefix创建一个mcode，格式：<Prefix>_<code>
	Prefix  string
	GinRoot *gin.Engine

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
	// 序列化函数
	MarshalFunc func(ctx *gin.Context, data interface{}) (string, []byte)

	// 不建议使用，自定义转换函数，修改本函数后会重新定义本函数后，其他的选项均不生效，需要使用者自己全部重写
	// 使用者如果需要注册自定义的函数格式使使用
	WrapFunc func(f interface{}) gin.HandlerFunc
}

// useDefault 补全配置
func (option *Option) useDefault() {
	if option.Prefix == "" {
		option.Prefix = "SERVICE"
	}

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

	if option.GinRoot == nil {
		option.GinRoot = gin.New()
		option.GinRoot.Use(gin.Recovery())
	}

	if option.MarshalFunc == nil {
		option.MarshalFunc = DefaultMarshalFunc
	}
}

// Wrapper 用于对请求返回结果进行封装的类
type Wrapper struct {
	// 错误码的前缀
	// 比如 错误码为1001，前缀为ANYPROJECT_ANYSERVICE_,那么返回给调用者的错误码(mcode)就为:ANYPROJECT_ANYSERVICE_1001
	mcodePrefix string
	// 服务名称
	serviceName string
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
	// 封装的gin.Engine对象
	ginRoot *gin.Engine
	// 序列化函数
	marshal MarshalFunc
	// 打开统计
	enableStat bool
}

// New 创建一个新的wrapper
func New(option *Option) *Wrapper {
	if option == nil {
		option = &Option{}
	}

	option.useDefault()

	w := &Wrapper{
		mcodePrefix: option.Prefix,
		snowSlide:   option.SnowSlide,
		wrapFunc:    option.WrapFunc,
		report:      option.Report,
		logger:      option.Logger,
		writeResult: option.WriteResult,
		logFunc:     option.LogFunc,
		ginRoot:     option.GinRoot,
		marshal:     option.MarshalFunc,
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

// HandlePprof 添加Pprof到处理列表
func (wrapper *Wrapper) HandlePprof() {
	debugPrintRoute(wrapper.logger, "PPROF", "/debug/pprof/.*", nil)
	ginpprof.Wrapper(wrapper.ginRoot)
}

// HandleStat 启动统计
func (wrapper *Wrapper) HandleStat() {
	wrapper.enableStat = true
	wrapper.Get("/debug/httpstat/delay", httpstat.StatDelay)
	wrapper.Get("/debug/httpstat/result", httpstat.StatResult)
}

// ServeHTTP http.Handler 实现
func (wrapper *Wrapper) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	wrapper.ginRoot.ServeHTTP(w, req)
}

// RunTLS 加密通信上运行HTTP服务
// 正常运行会堵塞
func (wrapper *Wrapper) RunTLS(address string, certFile string, keyFile string) error {
	return wrapper.ginRoot.RunTLS(address, certFile, keyFile)
}

// Run 运行HTTP服务
// 正常运行时会堵塞
func (wrapper *Wrapper) Run(address string) error {
	return wrapper.ginRoot.Run(address)
}

// GinEngine 返回gin.Engine对象
func (wrapper *Wrapper) GinEngine() *gin.Engine {
	return wrapper.ginRoot
}

// Group 构造一个组
func (wrapper *Wrapper) Group(path string) GroupWrapper {
	return &groupWrapper{
		wrapper:     wrapper,
		RouterGroup: wrapper.ginRoot.Group(path),
	}
}

func debugPrintRoute(logger *logrus.Logger, method, path string, f interface{}) {
	logger.WithField("path", path).Debugf("Handle %s", method)
}

// Handle Http通用请求注册
func (wrapper *Wrapper) Handle(method string, path string, f interface{}) {
	debugPrintRoute(wrapper.logger, method, path, f)
	wrapper.ginRoot.Handle(method, path, wrapper.wrapFunc(f))
}

// Get Http-GET请求注册
func (wrapper *Wrapper) Get(path string, f interface{}) {
	wrapper.Handle("GET", path, f)
}

// Patch Http-PATCH请求注册
func (wrapper *Wrapper) Patch(path string, f interface{}) {
	wrapper.Handle("PATCH", path, f)
}

// Post Http-POST请求注册
func (wrapper *Wrapper) Post(path string, f interface{}) {
	wrapper.Handle("POST", path, f)
}

// Put Http-PUT请求注册
func (wrapper *Wrapper) Put(path string, f interface{}) {
	wrapper.Handle("PUT", path, f)
}

// Options Http-OPTIONS请求注册
func (wrapper *Wrapper) Options(path string, f interface{}) {
	wrapper.Handle("OPTIONS", path, f)
}

// Head Http-HEADER请求注册
func (wrapper *Wrapper) Head(path string, f interface{}) {
	wrapper.Handle("HEAD", path, f)
}

// Delete Http-DELETE请求注册
func (wrapper *Wrapper) Delete(path string, f interface{}) {
	wrapper.Handle("DELETE", path, f)
}

// Any Http-所有请求注册
func (wrapper *Wrapper) Any(path string, f interface{}) {
	debugPrintRoute(wrapper.logger, "ANY", path, f)
	wrapper.ginRoot.Any(path, wrapper.wrapFunc(f))
}

// DefaultResponse 默认的返回
type DefaultResponse struct {
	Result    bool        `json:"result"`
	Mcode     string      `json:"mcode,omitempty"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
}
