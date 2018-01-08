package invoke

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/service/restful/code"
	"github.com/lworkltd/kits/service/monitor"
	"time"
	"strconv"
	"strings"
)

type Response struct {
	Result  bool            `json:"result"`
	Code    string          `json:"mcode"`
	Data    json.RawMessage `json:"data,omitempty"`
	Message string          `json:"message,omitemtpy"`
}

// ExtractHeader 解析包中的错误码(该封装已经达成共识)
// 即：{result:true,mcode:"<code>",data:{}}
func ExtractHeader(name string, invokeErr error, statusCode int, res *Response, out interface{}) code.Error {
	if statusCode == 0 {
		return code.NewMcode(
			fmt.Sprintf("INVOKE_FAILED"),
			invokeErr.Error(),
		)
	}

	if statusCode != http.StatusOK {
		return code.NewMcode(
			fmt.Sprintf("INVOKE_BAD_STATUS_%d", statusCode),
			fmt.Sprintf("service %s invoke failed,bad status code,%d", name, statusCode),
		)
	}

	if !res.Result {
		mcode := res.Code
		if mcode == "" {
			mcode = "INVOKE_FAILED_WITHOUT_MCODE"
		}
		return code.NewMcode(res.Code, res.Message)
	}

	if out == nil {
		return nil
	}

	err := json.Unmarshal(res.Data, out)
	if err != nil {
		return code.NewMcode(
			fmt.Sprintf("INVOKE_BAD_PAYLOAD"),
			err.Error(),
		)
	}

	return nil
}

func reportDataToMonitor(error code.Error, rsp *http.Response) {
	if monitor.EnableReportMonitor() == false || nil == rsp {			//rsp为nil时，已在client中错误上报
		return
	}
	timeNowMicrosecond := time.Now().UnixNano() / 1e3
	infc := rsp.Header.Get("Infc")
	tName := rsp.Header.Get("TName")
	endpoint := rsp.Header.Get("Endpoint")			//请求的IP:Port，或者一个domain:Port/domain
	tIP := endpoint
	endArray := strings.Split(endpoint, ":")
	if len(endArray) >= 2 {								//若有端口号，只保留IP或者domain
		tIP = endArray[0]
	}
	beginTimeMicrosecond,_ := strconv.ParseInt(rsp.Header.Get("BeginTime"), 10, 64)
	if nil == error {					//处理成功上报
		var succCountReport monitor.ReqSuccessCountDimension
		succCountReport.SName = monitor.GetCurrentServerName()
		succCountReport.SIP = monitor.GetCurrentServerIP()
		succCountReport.TName = tName
		succCountReport.TIP = tIP
		succCountReport.Infc = infc
		monitor.ReportReqSuccess(&succCountReport)

		if beginTimeMicrosecond > 0 {
			var succAvgTimeReport monitor.ReqSuccessAvgTimeDimension
			succAvgTimeReport.SName = monitor.GetCurrentServerName()
			succAvgTimeReport.SIP = monitor.GetCurrentServerIP()
			succAvgTimeReport.TName = tName
			succAvgTimeReport.TIP = tIP
			succAvgTimeReport.Infc = infc
			monitor.ReportSuccessAvgTime(&succAvgTimeReport, timeNowMicrosecond - beginTimeMicrosecond)		//耗时单位为微秒
		}
	} else {							//处理失败上报
		var failedCountReport monitor.ReqFailedCountDimension
		failedCountReport.SName = monitor.GetCurrentServerName()
		failedCountReport.TName = tName
		failedCountReport.TIP = tIP
		failedCountReport.Code = error.Mcode()
		failedCountReport.Infc = infc
		monitor.ReportReqFailed(&failedCountReport)

		if beginTimeMicrosecond > 0 {
			var failedAvgTimeReport monitor.ReqFailedAvgTimeDimension
			failedAvgTimeReport.SName = monitor.GetCurrentServerName()
			failedAvgTimeReport.SIP = monitor.GetCurrentServerIP()
			failedAvgTimeReport.TName = tName
			failedAvgTimeReport.TIP = tIP
			failedAvgTimeReport.Infc = infc
			monitor.ReportFailedAvgTime(&failedAvgTimeReport, timeNowMicrosecond - beginTimeMicrosecond) //耗时单位为微秒
		}
	}
}

// ExtractHttpResponse 解析标准http.Response为输出
func ExtractHttpResponse(name string, invokeErr error, rsp *http.Response, out interface{}) code.Error {
	var commonResp Response
	var errCode code.Error = nil
	if invokeErr == nil && rsp != nil {
		defer rsp.Body.Close()
	}

	statusCode := 0
	if rsp != nil {
		statusCode = rsp.StatusCode
	}

	if statusCode == http.StatusOK {
		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			errCode = code.NewMcode(fmt.Sprintf("INVOKE_READ_BODY_FAILED"),err.Error())
			reportDataToMonitor(errCode, rsp)
			return errCode
		}

		if len(body) == 0 {
			errCode = code.NewMcode(fmt.Sprintf("INVOKE_EMPTY_BODY"),err.Error())
			reportDataToMonitor(errCode, rsp)
			return errCode
		}
		logrus.Debug("Invoke return Body", string(body))
		err = json.Unmarshal(body, &commonResp)
		if err != nil {
			errCode = code.NewMcode(fmt.Sprintf("INVOKE_PARSE_COMMON_RSP_FAILED"),err.Error())
			reportDataToMonitor(errCode, rsp)
			return errCode
		}
	}

	errCode = ExtractHeader(name, invokeErr, statusCode, &commonResp, out)
	reportDataToMonitor(errCode, rsp)
	return errCode
}
