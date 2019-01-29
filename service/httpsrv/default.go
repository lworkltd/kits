package httpsrv

import (
	"fmt"
	"net/http"
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

// DefaultWrap 默认的函数转换
func (wrapper *Wrapper) DefaultWrap(fx interface{}) gin.HandlerFunc {
	f := fx.(func(ginCtx *gin.Context) (interface{}, code.Error))
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
					fmt.Println("Panic reason:", r)
					fmt.Println(string(debug.Stack()))
					cerr = CerrIntervalError
				}
			}

			wrapper.writeResult(httpCtx, wrapper.mcodePrefix, data, cerr)
			wrapper.logFunc(wrapper.logger, httpCtx, since, data, cerr)
		}()

		// 过载保护
		if wrapper.snowSlide != nil {
			cerr = wrapper.snowSlide.Check(httpCtx)
			if cerr == nil {
				data, cerr = f(httpCtx)
			}
		} else {
			data, cerr = f(httpCtx)
		}

		wrapper.report(cerr, httpCtx, since)
	}
}

// LogFunc 日志处理函数
// logger logrus打印日志的接口
// ctx 本次请求的上下文(gin)
// cerr 本次请求的错误
// data 返回的数据
// since 开始处理请求的时间
type LogFunc func(logger *logrus.Logger, ctx *gin.Context, since time.Time, data interface{}, cerr code.Error)

// DefaultLogFunc 默认的日志打印函数
func DefaultLogFunc(logger *logrus.Logger, ctx *gin.Context, since time.Time, data interface{}, cerr code.Error) {
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
type ReportFunc func(err code.Error, ctx *gin.Context, since time.Time)

// DefaultReport 上报处理请求结果到Monitor，registPath为注册路径
func DefaultReport(err code.Error, httpCtx *gin.Context, since time.Time) {
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
type WriteResultFunc func(ctx *gin.Context, prefix string, data interface{}, cerr code.Error)

// DefaultWriteResultFunc 默认写结果
func DefaultWriteResultFunc(ctx *gin.Context, prefix string, data interface{}, cerr code.Error) {
	// 错误的返回
	if cerr != nil {
		if cerr.Mcode() != "" {
			ctx.JSON(http.StatusOK, map[string]interface{}{
				"result":    false,
				"mcode":     cerr.Mcode(),
				"message":   cerr.Message(),
				"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
			})

		} else {
			mcode := fmt.Sprintf("%s_%d", prefix, cerr.Code())
			ctx.JSON(http.StatusOK, map[string]interface{}{
				"result":    false,
				"mcode":     mcode,
				"message":   cerr.Message(),
				"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
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
		ctx.JSON(http.StatusOK, resp)
	}
}
