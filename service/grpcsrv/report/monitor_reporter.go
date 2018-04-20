package report

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/lworkltd/kits/service/monitor"
)

// RpcReporter 上报
type RpcReporter interface {
	Report(reqInterface, reqService, fromHost string, result string, deplay time.Duration)
}

// MonitorReporter 基于 service/monitor 的上报
type MonitorReporter struct {
}

// Report 上报
func (reporter *MonitorReporter) Report(reqInterface, reqService, fromHost string, result string, delay time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			fmt.Println(debug.Stack())
		}
	}()

	isSucc := result == ""
	infc := fmt.Sprintf("ACTIVE_GRPC_%s", reqInterface)
	localIp := monitor.GetCurrentServerIP()
	delayMs := int64(delay / time.Millisecond)
	if isSucc {
		monitor.ReportReqSuccess(&monitor.ReqSuccessCountDimension{
			SName: "",
			SIP:   fromHost,
			TName: reqService,
			TIP:   localIp,
			Infc:  infc,
		})
		monitor.ReportSuccessAvgTime(&monitor.ReqSuccessAvgTimeDimension{
			SName: "",
			SIP:   fromHost,
			TName: reqService,
			TIP:   localIp,
			Infc:  infc,
		}, delayMs)
	} else {
		monitor.ReportReqFailed(&monitor.ReqFailedCountDimension{
			SName: "",
			TName: reqService,
			TIP:   localIp,
			Code:  result,
			Infc:  infc,
		})
		monitor.ReportSuccessAvgTime(&monitor.ReqSuccessAvgTimeDimension{
			SName: "",
			SIP:   fromHost,
			TName: reqService,
			TIP:   localIp,
			Infc:  infc,
		}, delayMs)
	}
}
