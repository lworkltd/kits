package httpsrv

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/monitor"
	"github.com/lworkltd/kits/service/restful/code"
)

// ReportFunc 上报处理函数
type ReportFunc func(err code.Error, httpCtx *gin.Context, beginTime time.Time)

//上报处理请求结果到Monitor，registPath为注册路径
func defaultReportProcessResultToMonitor(err code.Error, httpCtx *gin.Context, beginTime time.Time) {
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
