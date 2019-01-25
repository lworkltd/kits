package invoke

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/lworkltd/kits/service/monitor"
	"github.com/lworkltd/kits/service/restful/code"
)

type Response struct {
	Result  bool            `json:"result"`
	Code    string          `json:"mcode,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
	Message string          `json:"message,omitempty"`
}

var (
	MCODE_INVOKE_TIMEOUT = "INVOKE_TIMEOUT"
	MCODE_INVOKE_FAILED  = "INVOKE_FAILED"
)

func reportDataToMonitor(error code.Error, rsp *http.Response) {
	if monitor.EnableReportMonitor() == false || nil == rsp { //rsp为nil时，已在client中错误上报
		return
	}
	timeNowMicrosecond := time.Now().UnixNano() / 1e3
	infc := rsp.Header.Get("Infc")
	tName := rsp.Header.Get("TName")
	endpoint := rsp.Header.Get("Endpoint") //请求的IP:Port，或者一个domain:Port/domain
	tIP := endpoint
	endArray := strings.Split(endpoint, ":")
	if len(endArray) >= 2 { //若有端口号，只保留IP或者domain
		tIP = endArray[0]
	}
	beginTimeMicrosecond, _ := strconv.ParseInt(rsp.Header.Get("BeginTime"), 10, 64)
	if nil == error { //处理成功上报
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
			monitor.ReportSuccessAvgTime(&succAvgTimeReport, timeNowMicrosecond-beginTimeMicrosecond) //耗时单位为微秒
		}
	} else { //处理失败上报
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
			monitor.ReportFailedAvgTime(&failedAvgTimeReport, timeNowMicrosecond-beginTimeMicrosecond) //耗时单位为微秒
		}
	}
}

// ExtractHeader 解析包中的错误码(该封装已经达成共识)
// 即：{result:true,mcode:"<code>",data:{}}
func extractHeader(invokeErr error, statusCode int, res *Response, out interface{}) code.Error {
	if statusCode == 0 {
		// HTTP 调用过程出错
		urlErr, ok := invokeErr.(*url.Error)
		if ok {
			// 超时错误
			if urlErr.Timeout() {
				return code.NewMcode(MCODE_INVOKE_TIMEOUT, "http timeout")
			}

			if netErr, ok := urlErr.Err.(net.Error); ok {
				if netErr.Timeout() {
					return code.NewMcode(MCODE_INVOKE_TIMEOUT, "network timeout")
				}

				if netOpErr, ok := netErr.(*net.OpError); ok {
					if netOpErr.Timeout() {
						return code.NewMcode(MCODE_INVOKE_TIMEOUT, "network operation timeout")
					}

					if sysCallErr, ok := netOpErr.Err.(*os.SyscallError); ok {
						if sysCallErr.Syscall == "connectx" {
							return code.NewMcodef(MCODE_INVOKE_FAILED,
								"connect failed,addr=%v", netOpErr.Addr.String())
						}
						return code.NewMcode(MCODE_INVOKE_FAILED, "invoke failed for net problem")
					}
				}
			}

			return code.NewMcodef(MCODE_INVOKE_FAILED, "invoke error,%v", urlErr)
		}

		// 超时熔断
		if invokeErr == hystrix.ErrTimeout {
			return code.NewMcode(MCODE_INVOKE_TIMEOUT, invokeErr.Error())
		}

		// 熔断开启
		if invokeErr == hystrix.ErrCircuitOpen {
			return code.NewMcode(MCODE_INVOKE_FAILED, invokeErr.Error())
		}

		// 过载熔断
		if invokeErr == hystrix.ErrMaxConcurrency {
			return code.NewMcode(MCODE_INVOKE_FAILED, invokeErr.Error())
		}

		// 其他错误
		return code.NewMcode(
			fmt.Sprintf(MCODE_INVOKE_FAILED),
			invokeErr.Error(),
		)
	}

	// 返回状态出错
	if statusCode != http.StatusOK {
		return code.NewMcode(
			MCODE_INVOKE_FAILED,
			fmt.Sprintf("http status: %d", statusCode),
		)
	}

	// 处理结果出错
	if !res.Result {
		mcode := res.Code
		if mcode == "" {
			mcode = MCODE_INVOKE_FAILED
		}
		return code.NewMcode(mcode, res.Message)
	}

	// 无需解析结果
	if out == nil {
		return nil
	}

	err := json.Unmarshal(res.Data, out)
	if err != nil {
		return code.NewMcode(
			MCODE_INVOKE_FAILED,
			"parse return json payload failed",
		)
	}

	return nil
}

// ExtractHttpResponse 解析标准http.Response为输出
func extractHttpResponse(invokeErr error, rsp *http.Response, out interface{}) code.Error {
	var commonResp Response
	var errCode code.Error

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
			errCode = code.NewMcode(MCODE_INVOKE_FAILED, "read response body failed")
			reportDataToMonitor(errCode, rsp)
			return errCode
		}

		if len(body) == 0 {
			errCode = code.NewMcode(MCODE_INVOKE_FAILED, "return json body empty")
			reportDataToMonitor(errCode, rsp)
			return errCode
		}

		err = json.Unmarshal(body, &commonResp)
		if err != nil {
			errCode = code.NewMcode(MCODE_INVOKE_FAILED, "return json body error")
			reportDataToMonitor(errCode, rsp)
			return errCode
		}
	}

	errCode = extractHeader(invokeErr, statusCode, &commonResp, out)
	reportDataToMonitor(errCode, rsp)
	return errCode
}

// Result 获取结果
// out json中data反序列化到out中
func (client *client) Result(out interface{}) code.Error {
	rsp, err := client.response()
	if err != nil {
		return extractHttpResponse(err, rsp, out)
	}
	return nil
}
