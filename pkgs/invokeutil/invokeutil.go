package invokeutil

import (
	"encoding/json"
	"fmt"
	"github.com/lvhuat/kits/service/restful/code"
	"net/http"
	"strings"
)

type Response struct {
	Result  bool            `json:"result"`
	Code    string          `json:"mcode"`
	Data    json.RawMessage `json:"data,omitempty"`
	Message string          `json:"message,omitemtpy"`
}

// 解析包中的错误码(该封装已经达成共识)
// 即：{result:true,mcode:"<code>",data:{}}
func Unpkg(name string, invokeErr error, statusCode int, res *Response, out interface{}) code.Error {
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
