package invoke

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/service/restful/code"
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


// ExtractHttpResponse 解析标准http.Response为输出
func ExtractHttpResponse(name string, invokeErr error, rsp *http.Response, out interface{}) code.Error {
	var commonResp Response
	statusCode := 0
	if rsp != nil {
		statusCode = rsp.StatusCode
	}

	if statusCode == http.StatusOK {
		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return code.NewMcode(
				fmt.Sprintf("INVOKE_READ_BODY_FAILED"),
				err.Error(),
			)
		}

		if len(body) == 0 {
			return code.NewMcode(
				fmt.Sprintf("INVOKE_EMPTY_BODY"),
				err.Error(),
			)
		}
		logrus.Debug("Invoke return Body", string(body))
		err = json.Unmarshal(body, &commonResp)
		if err != nil {
			return code.NewMcode(
				fmt.Sprintf("INVOKE_PARSE_COMMON_RSP_FAILED"),
				err.Error(),
			)
		}
	}

	return ExtractHeader(name, invokeErr, statusCode, &commonResp, out)
}
