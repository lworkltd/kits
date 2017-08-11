package invoke

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
			fmt.Sprintf("%s_BAD_INVOKE", strings.ToUpper(name)),
			invokeErr.Error(),
		)
	}

	if statusCode != http.StatusOK {
		return code.NewMcode(
			fmt.Sprintf("%s_BAD_STATUS_%d", strings.ToUpper(name), statusCode),
			fmt.Sprintf("service %s invoke failed,bad status code,%d", name, statusCode),
		)
	}

	if !res.Result {
		return code.NewMcode(res.Code, res.Message)
	}

	err := json.Unmarshal(res.Data, out)
	if err != nil {
		return code.NewMcode(
			fmt.Sprintf("%s_BAD_RES_PAYLOAD", strings.ToUpper(name)),
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
	return ExtractHeader(name, invokeErr, statusCode, &commonResp, out)
}
