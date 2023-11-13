package wrap

import (
	"fmt"
	"log"
	"os"

	"sync"

	"net/http"
	"runtime/debug"
	"time"

	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/context"
	"github.com/lworkltd/kits/service/monitor"
	"github.com/lworkltd/kits/service/restful/code"
	logutils "github.com/lworkltd/kits/utils/log"
	"github.com/sirupsen/logrus"
)

var (
	// DefaultSnowSlideLimit  默认过载保护
	DefaultSnowSlideLimit int32 = 20000
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
	serviceId        string
	serviceLogLevel  logrus.Level
	serviceLogWriter io.Writer

	snowSlide *SnowSlide

	logFn func(entry *logrus.Entry, level logrus.Level, msg string)
}

type Option struct {
	Prefix      string
	Mode        string
	LogLevel    string
	LogFilePath string

	// SnowSlideLimit 过载保护数，<=0 时，不受限，大于0时，限制秒内最大请求数
	SnowSlideLimit int32
}

// New 创建一个新的wrapper
func New(option *Option) *Wrapper {
	// 设置日志输出IO流，若未配置使用os.Stderr
	logWriter := os.Stderr
	if "" != option.LogFilePath {
		file, err := os.OpenFile(option.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if nil != err {
			panic(fmt.Errorf("Open log file failed, err:%v, log file path:%v", err, option.LogFilePath))
		} else {
			logWriter = file
		}
	}

	// 设置日志等级，若未配置，使用logrus.InfoLevel
	logLevel := logrus.InfoLevel
	if option.LogLevel != "" {
		logLevelParse, err := logrus.ParseLevel(option.LogLevel)
		if err != nil {
			panic(fmt.Errorf("cannot parse logger level %s", option.LogLevel))
		} else {
			logLevel = logLevelParse
		}
	}

	//  初始化过载保护
	var snowSlide *SnowSlide
	if option.SnowSlideLimit <= 0 {
		option.SnowSlideLimit = DefaultSnowSlideLimit
	}
	snowSlide = &SnowSlide{
		LimitCnt: option.SnowSlideLimit,
		Service:  option.Prefix,
	}

	w := &Wrapper{
		mcodePrefix:      option.Prefix,
		serviceLogLevel:  logLevel,
		serviceLogWriter: logWriter,
		pool: sync.Pool{
			New: func() interface{} {
				return new(Response)
			},
		},
		snowSlide: snowSlide,
		logFn: func(entry *logrus.Entry, level logrus.Level, msg string) {
			entry.Log(level, msg)
		},
	}

	return w
}

// WrappedFunc 是用于封装GIN HTTP接口返回为通用接口的函数定义
type WrappedFunc func(srvContext context.Context, ctx *gin.Context) (interface{}, code.Error)

type HttpServer interface {
	Handle(string, string, ...gin.HandlerFunc) gin.IRoutes
}

//上报处理请求结果到Monitor，registPath为注册路径
func reportProcessResultToMonitor(err code.Error, httpCtx *gin.Context, beginTime time.Time, registPath string) {
	if nil == httpCtx || false == monitor.EnableReportMonitor() {
		return
	}
	timeNow := time.Now()
	infc := "PASSIVE_" + httpCtx.Request.Method + "_" + registPath //PASSIVE表示被调, httpCtx.Request.URL.Path为实际请求的路径
	addrs := strings.Split(httpCtx.Request.RemoteAddr, ":")        //httpCtx.Request.RemoteAddr, 例如：118.112.177.203:58425
	sIP := ""
	if len(addrs) == 2 && monitor.IsInnerIPv4(addrs[0]) {
		sIP = addrs[0] //若远端IP为内网IP则取值，公网IP请求过多会导致Monitor数据标签量太大
	}
	if nil == err { //处理成功
		//请求失败，上报失败计数和失败平均耗时
		timeNow := time.Now()
		var succCountReport monitor.ReqSuccessCountDimension
		succCountReport.SName = ""
		succCountReport.SIP = sIP
		succCountReport.TName = monitor.GetCurrentServerName()
		succCountReport.TIP = monitor.GetCurrentServerIP()
		succCountReport.Infc = infc
		monitor.ReportReqSuccess(&succCountReport)

		var succAvgTimeReport monitor.ReqSuccessAvgTimeDimension
		succAvgTimeReport.SName = ""
		succAvgTimeReport.SIP = sIP
		succAvgTimeReport.TName = monitor.GetCurrentServerName()
		succAvgTimeReport.TIP = monitor.GetCurrentServerIP()
		succAvgTimeReport.Infc = infc
		monitor.ReportSuccessAvgTime(&succAvgTimeReport, (timeNow.UnixNano()-beginTime.UnixNano())/1e3) //耗时单位为微秒
	} else { //处理失败
		var failedCountReport monitor.ReqFailedCountDimension
		failedCountReport.SName = ""
		failedCountReport.TName = monitor.GetCurrentServerName()
		failedCountReport.TIP = monitor.GetCurrentServerIP()
		failedCountReport.Code = err.Mcode()
		failedCountReport.Infc = infc
		monitor.ReportReqFailed(&failedCountReport)

		var failedAvgTimeReport monitor.ReqFailedAvgTimeDimension
		failedAvgTimeReport.SName = ""
		failedAvgTimeReport.SIP = sIP
		failedAvgTimeReport.TName = monitor.GetCurrentServerName()
		failedAvgTimeReport.TIP = monitor.GetCurrentServerIP()
		failedAvgTimeReport.Infc = infc
		monitor.ReportFailedAvgTime(&failedAvgTimeReport, (timeNow.UnixNano()-beginTime.UnixNano())/1e3) //耗时单位为微秒
	}
}

// Wrap 为gin的回调接口增加了固定的返回值，当程序收到处理结果的时候会将返回值封装一层再发送到网络, registPath为注册路径
func (wrapper *Wrapper) Wrap(f WrappedFunc, registPath string) gin.HandlerFunc {
	return func(httpCtx *gin.Context) {
		Prefix := wrapper.mcodePrefix // 错误码前缀
		logger := logrus.New()
		// 设置日志输出IO流
		logger.Out = wrapper.serviceLogWriter
		// 设置日志等级
		logger.Level = wrapper.serviceLogLevel

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

			l := serviceCtx.WithFields(logrus.Fields{
				"method": httpCtx.Request.Method,
				"path":   httpCtx.Request.URL.Path,
				"delay":  time.Since(since),
			})

			// 错误的返回
			if cerr != nil {
				if cerr.Mcode() != "" {
					httpCtx.JSON(http.StatusOK, map[string]interface{}{
						"result":    false,
						"mcode":     cerr.Mcode(),
						"message":   cerr.Message(),
						"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
					})

					l = l.WithFields(logrus.Fields{
						"mcode": cerr.Mcode(),
					})
				} else {
					mcode := fmt.Sprintf("%s_%d", Prefix, cerr.Code())
					httpCtx.JSON(http.StatusOK, map[string]interface{}{
						"result":    false,
						"mcode":     mcode,
						"message":   cerr.Message(),
						"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
					})

					l = l.WithFields(logrus.Fields{
						"mcode": mcode,
					})
				}
			} else {
				resp := map[string]interface{}{
					"result":    true,
					"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
				}
				if data != nil {
					resp["data"] = data
				}
				httpCtx.JSON(http.StatusOK, resp)
			}

			var level logrus.Level
			var msg string

			if cerr != nil {
				l = l.WithFields(logrus.Fields{
					"message": cerr.Message(),
				})
				msg = "HTTP request failed"
				level = logrus.ErrorLevel
			} else {
				msg = "HTTP request done"
				level = logrus.InfoLevel
			}

			wrapper.logFn(l, level, msg)
		}()

		// 过载保护
		if wrapper.snowSlide != nil {
			cerr = wrapper.snowSlide.Check()
			if cerr == nil {
				data, cerr = f(serviceCtx, httpCtx)
			}
		} else {
			data, cerr = f(serviceCtx, httpCtx)
		}

		reportProcessResultToMonitor(cerr, httpCtx, since, registPath)
	}
}

func (wrapper *Wrapper) Handle(method string, srv HttpServer, path string, f WrappedFunc) {
	registPath := srv.(*gin.RouterGroup).BasePath() + path
	srv.Handle(method, path, wrapper.Wrap(f, registPath))
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

func (wrapper *Wrapper) SetLogger(logFn func(entry *logrus.Entry, level logrus.Level, msg string)) {
	wrapper.logFn = logFn
}
