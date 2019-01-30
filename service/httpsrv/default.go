package httpsrv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/monitor"
	"github.com/lworkltd/kits/service/restful/code"
	"github.com/sirupsen/logrus"
)

var (
	// CerrIntervalError 当出现不可辨认的异常时将返回此错误
	CerrIntervalError = code.NewMcode("SERVICE_INTERVAL_ERROR", "Service internal error")
)

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

// wrapperInterfaceCodeerror 注册函数类型为： func(ginCtx *gin.Context) (interface{}, code.Error)
func (wrapper *Wrapper) wrapperDataCodeerror(f func(ginCtx *gin.Context) (interface{}, code.Error)) func(httpCtx *gin.Context) {
	return func(httpCtx *gin.Context) {
		since := time.Now()
		var (
			data interface{}
			cerr code.Error
		)

		defer func() {
			if r := recover(); r != nil {
				if codeErr, ok := r.(code.Error); ok { // 接受上层业务通过Panic(code.Error)的方式来返回错误,但是请慎用这种方式
					cerr = codeErr
				} else {
					DoPanic(r)
					cerr = CerrIntervalError
				}
			}

			wrapper.writeResult(httpCtx, wrapper.marshal, http.StatusOK, wrapper.mcodePrefix, data, cerr)
			wrapper.logFunc(wrapper.logger, httpCtx, http.StatusOK, since, data, cerr)
			wrapper.report(cerr, httpCtx, http.StatusOK, since)
		}()

		// 过载保护
		if wrapper.snowSlide != nil {
			cerr = wrapper.snowSlide.Check(httpCtx)
			if cerr != nil {
				return
			}
		}

		data, cerr = f(httpCtx)
	}
}

// wrapperInterfaceError 注册函数类型为: func(ginCtx *gin.Context) (interface{}, error)
func (wrapper *Wrapper) wrapperDataError(f func(ginCtx *gin.Context) (interface{}, error)) func(httpCtx *gin.Context) {
	return func(httpCtx *gin.Context) {
		since := time.Now()
		var (
			data interface{}
			cerr code.Error
		)

		defer func() {
			if r := recover(); r != nil {
				if codeErr, ok := r.(code.Error); ok { // 接受上层业务通过Panic(code.Error)的方式来返回错误,但是请慎用这种方式
					cerr = codeErr
				} else {
					DoPanic(r)
					cerr = CerrIntervalError
				}
			}

			wrapper.writeResult(httpCtx, wrapper.marshal, http.StatusOK, wrapper.mcodePrefix, data, cerr)
			wrapper.report(cerr, httpCtx, http.StatusOK, since)
			wrapper.logFunc(wrapper.logger, httpCtx, http.StatusOK, since, data, cerr)
		}()

		// 过载保护
		if wrapper.snowSlide != nil {
			cerr = wrapper.snowSlide.Check(httpCtx)
			if cerr != nil {
				return
			}
		}

		d, err := f(httpCtx)
		data = d
		if err != nil {
			panic(err)
		}
	}
}

// wrapperNormal 注册函数类型为：func(ginCtx *gin.Context)
// 该注册类型不向IO写入数据
func (wrapper *Wrapper) wrapperNormal(f func(ginCtx *gin.Context)) func(httpCtx *gin.Context) {
	return func(httpCtx *gin.Context) {
		since := time.Now()
		var (
			data interface{}
			cerr code.Error
		)

		defer func() {
			if r := recover(); r != nil {
				if codeErr, ok := r.(code.Error); ok { // 接受上层业务通过Panic(code.Error)的方式来返回错误,但是请慎用这种方式
					cerr = codeErr
				} else {
					DoPanic(r)
					cerr = CerrIntervalError
				}
			}

			wrapper.report(cerr, httpCtx, http.StatusOK, since)
			wrapper.logFunc(wrapper.logger, httpCtx, http.StatusOK, since, data, cerr)
		}()

		// 过载保护
		if wrapper.snowSlide != nil {
			cerr = wrapper.snowSlide.Check(httpCtx)
			if cerr != nil {
				return
			}
		}

		f(httpCtx)
	}
}

// wrapperError 注册函数类型为：func(ginCtx *gin.Context) error
// 该注册函数认为没有数据需要返回
func (wrapper *Wrapper) wrapperError(f func(ginCtx *gin.Context) error) func(httpCtx *gin.Context) {
	return func(httpCtx *gin.Context) {
		since := time.Now()
		var (
			cerr code.Error
		)

		defer func() {
			if r := recover(); r != nil {
				if codeErr, ok := r.(code.Error); ok { // 接受上层业务通过Panic(code.Error)的方式来返回错误,但是请慎用这种方式
					cerr = codeErr
				} else {
					DoPanic(r)
					cerr = CerrIntervalError
				}
			}

			wrapper.writeResult(httpCtx, wrapper.marshal, http.StatusOK, wrapper.mcodePrefix, nil, cerr)
			wrapper.report(cerr, httpCtx, http.StatusOK, since)
			wrapper.logFunc(wrapper.logger, httpCtx, http.StatusOK, since, nil, cerr)
		}()

		// 过载保护
		if wrapper.snowSlide != nil {
			cerr = wrapper.snowSlide.Check(httpCtx)
			if cerr != nil {
				return
			}
		}

		err := f(httpCtx)
		if err != nil {
			panic(err)
		}
	}
}

// wrapperError 注册函数类型为：func(ginCtx *gin.Context) code.Error
// 该注册函数认为没有数据需要返回
func (wrapper *Wrapper) wrapperCodeError(f func(ginCtx *gin.Context) code.Error) func(httpCtx *gin.Context) {
	return func(httpCtx *gin.Context) {
		since := time.Now()
		var (
			cerr code.Error
		)

		defer func() {
			if r := recover(); r != nil {
				if codeErr, ok := r.(code.Error); ok { // 接受上层业务通过Panic(code.Error)的方式来返回错误,但是请慎用这种方式
					cerr = codeErr
				} else {
					DoPanic(r)
					cerr = CerrIntervalError
				}
			}

			wrapper.writeResult(httpCtx, wrapper.marshal, http.StatusOK, wrapper.mcodePrefix, nil, cerr)
			wrapper.report(cerr, httpCtx, http.StatusOK, since)
			wrapper.logFunc(wrapper.logger, httpCtx, http.StatusOK, since, nil, cerr)
		}()

		// 过载保护
		if wrapper.snowSlide != nil {
			cerr = wrapper.snowSlide.Check(httpCtx)
			if cerr != nil {
				return
			}
		}

		cerr = f(httpCtx)
	}
}

// wrapperNoWrapperStatus 注册函数类型为：func(ginCtx *gin.Context)  (NoWrapperResponse, int)
// 该注册函数认为返回数据不需要额外封装，直接序列化后写入IO即可
// 返回的第一参数：无额外封装数据
// 返回的第二参数：http状态码，>= 400 表示失败
func (wrapper *Wrapper) wrapperNoWrapperStatus(f func(ginCtx *gin.Context) (NoWrapperResponse, int)) func(httpCtx *gin.Context) {
	return func(httpCtx *gin.Context) {
		since := time.Now()
		var (
			data       interface{}
			status     int
			statusData string
		)

		defer func() {
			if r := recover(); r != nil {
				DoPanic(r)
				status = http.StatusInternalServerError
				statusData = "ServiceIntervalError"
			}

			if status == 0 {
				status = http.StatusOK
			}

			contentType, body := wrapper.marshal(httpCtx, data)
			httpCtx.Data(status, contentType, body)

			var cerr code.Error
			if status >= 400 {
				cerr = code.NewMcode(fmt.Sprintf("HTTP_STATUS_%d", status), statusData)
			}
			wrapper.report(cerr, httpCtx, status, since)
			wrapper.logFunc(wrapper.logger, httpCtx, status, since, nil, cerr)
		}()

		// 过载保护
		if wrapper.snowSlide != nil {
			cerr := wrapper.snowSlide.Check(httpCtx)
			if cerr != nil {
				status = http.StatusTooManyRequests
				data = "too many request"
				return
			}
		}

		data, status = f(httpCtx)
	}
}

// wrapperNoWrapperStatus 注册函数类型为：func(ginCtx *gin.Context) NoWrapperResponse
// 该注册函数认为返回数据不需要额外封装，直接序列化后写入IO即可，除拦截器和异常外，一律返回200
// 返回的第一参数：无额外封装数据
func (wrapper *Wrapper) wrapperNoWrapper(f func(ginCtx *gin.Context) NoWrapperResponse) func(httpCtx *gin.Context) {
	return func(httpCtx *gin.Context) {
		since := time.Now()
		var (
			data       interface{}
			status     = http.StatusOK
			statusData string
		)

		defer func() {
			if r := recover(); r != nil {
				DoPanic(r)
				status = http.StatusInternalServerError
				statusData = "ServiceIntervalError"
			}

			contentType, body := wrapper.marshal(httpCtx, data)
			httpCtx.Data(status, contentType, body)

			var cerr code.Error
			if status != http.StatusOK {
				cerr = code.NewMcode(fmt.Sprintf("HTTP_STATUS_%d", status), statusData)
			}
			wrapper.report(cerr, httpCtx, status, since)
			wrapper.logFunc(wrapper.logger, httpCtx, status, since, nil, cerr)
		}()

		// 过载保护
		if wrapper.snowSlide != nil {
			cerr := wrapper.snowSlide.Check(httpCtx)
			if cerr != nil {
				status = http.StatusTooManyRequests
				data = "too many request"
				return
			}
		}

		data = f(httpCtx)
	}
}

// NoWrapperResponse 返回数据不需要额外封装的接口
// 如果注册函数返回类型实现func(ginCtx *gin.Context) NoWrapperPlease
// 则返回数据将不做额外封装，直接将数据序列化后返回
type NoWrapperResponse interface {
	NoWrapperPlease()
}

// DefaultWrap 默认的函数转换
func (wrapper *Wrapper) DefaultWrap(fx interface{}) gin.HandlerFunc {
	switch fx.(type) {
	case func(ginCtx *gin.Context) (interface{}, code.Error):
		return wrapper.wrapperDataCodeerror(fx.(func(ginCtx *gin.Context) (interface{}, code.Error)))
	case func(ginCtx *gin.Context) (interface{}, error):
		return wrapper.wrapperDataError(fx.(func(ginCtx *gin.Context) (interface{}, error)))
	case func(ginCtx *gin.Context):
		return wrapper.wrapperNormal(fx.(func(ginCtx *gin.Context)))
	case func(ginCtx *gin.Context) error:
		return wrapper.wrapperError(fx.(func(ginCtx *gin.Context) error))
	case func(ginCtx *gin.Context) code.Error:
		return wrapper.wrapperCodeError(fx.(func(ginCtx *gin.Context) code.Error))
	case func(ginCtx *gin.Context) (NoWrapperResponse, int):
		return wrapper.wrapperNoWrapperStatus(fx.(func(ginCtx *gin.Context) (NoWrapperResponse, int)))
	case func(ginCtx *gin.Context) NoWrapperResponse:
		return wrapper.wrapperNoWrapper(fx.(func(ginCtx *gin.Context) NoWrapperResponse))
	}
	panic(fmt.Sprintf("Unsupport register function type:%v", reflect.TypeOf(fx).String()))
}

// DoPanic 拦截到panic时执行
var DoPanic = func(r interface{}) {
	fmt.Println("Panic reason:", r)
	fmt.Println(string(debug.Stack()))
}

// LogFunc 日志处理函数
// logger logrus打印日志的接口
// ctx 本次请求的上下文(gin)
// cerr 本次请求的错误
// data 返回的数据
// since 开始处理请求的时间
type LogFunc func(logger *logrus.Logger, ctx *gin.Context, status int, since time.Time, data interface{}, cerr code.Error)

// DefaultLogFunc 默认的日志打印函数
func DefaultLogFunc(logger *logrus.Logger, ctx *gin.Context, status int, since time.Time, data interface{}, cerr code.Error) {
	l := logger.WithFields(logrus.Fields{
		"method": ctx.Request.Method,
		"path":   ctx.Request.URL.Path,
		"delay":  time.Since(since),
	})

	if cerr != nil {
		l = l.WithFields(logrus.Fields{
			"mcode": cerr.Mcode(),
		})
		l.WithFields(logrus.Fields{
			"message": cerr.Message(),
		}).Error("HTTP Failed")
	} else {

		l.Info("HTTP OK")
	}
}

// ReportFunc 上报处理函数
// err 结果错误
// ctx 本次请求的上下文(gin)
// since 开始处理请求的时间
type ReportFunc func(err code.Error, ctx *gin.Context, status int, since time.Time)

// DefaultReport 上报处理请求结果到Monitor，registPath为注册路径
func DefaultReport(err code.Error, httpCtx *gin.Context, status int, since time.Time) {
	if nil == httpCtx || false == monitor.EnableReportMonitor() {
		return
	}
	timeNow := time.Now()
	infc := "PASSIVE_" + httpCtx.Request.Method + "_" + httpCtx.Request.URL.Path //PASSIVE表示被调, httpCtx.Request.URL.Path为实际请求的路径
	addrs := strings.Split(httpCtx.Request.RemoteAddr, ":")                      //httpCtx.Request.RemoteAddr, 例如：118.112.177.203:58425
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
		monitor.ReportSuccessAvgTime(&succAvgTimeReport, (timeNow.UnixNano()-since.UnixNano())/1e3) //耗时单位为微秒
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
		monitor.ReportFailedAvgTime(&failedAvgTimeReport, (timeNow.UnixNano()-since.UnixNano())/1e3) //耗时单位为微秒
	}
}

// WriteResultFunc 写返回的函数
// ctx 本次请求的上下文(gin)
// prefix 错误码前缀
// cerr 处理错误
// data 结果数据
type WriteResultFunc func(ctx *gin.Context, marshal MarshalFunc, status int, prefix string, data interface{}, cerr code.Error)

// DefaultWriteResultFunc 默认写结果
func DefaultWriteResultFunc(ctx *gin.Context, marshal MarshalFunc, status int, prefix string, data interface{}, cerr code.Error) {
	var resp map[string]interface{}
	// 错误的返回
	if cerr != nil {
		if cerr.Mcode() != "" {
			resp = map[string]interface{}{
				"result":    false,
				"mcode":     cerr.Mcode(),
				"message":   cerr.Message(),
				"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
			}
		} else {
			mcode := fmt.Sprintf("%s_%d", prefix, cerr.Code())
			resp = map[string]interface{}{
				"result":    false,
				"mcode":     mcode,
				"message":   cerr.Message(),
				"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
			}
		}

		if data != nil {
			resp["data"] = data
		}
	} else {
		resp = map[string]interface{}{
			"result":    true,
			"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
		}
		if data != nil {
			resp["data"] = data
		}
	}

	contentType, body := marshal(ctx, resp)
	ctx.Data(status, contentType, body)
}

// MarshalFunc 序列化的函数
// data 需要序列化的数据
// 返回参数1：Content-Type的值
// 返回参数2：body数据
type MarshalFunc func(ctx *gin.Context, data interface{}) (string, []byte)

// DefaultMarshalFunc 默认的序列化函数
var DefaultMarshalFunc = JSONMarshalFunc

// JSONMarshalFunc JSON的序列化函数
func JSONMarshalFunc(ctx *gin.Context, data interface{}) (string, []byte) {
	body := []byte{}
	if data != nil {
		body, _ = json.Marshal(data)
	}

	return "application/json", body
}
