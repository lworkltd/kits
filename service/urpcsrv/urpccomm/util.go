package urpccomm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lworkltd/kits/service/restful/code"
)

// NotIntCodeError 非数字类型错误，不以数字为结尾的，即非业务错误
const NotIntCodeError = -11

// GrpcError grpc调用错误
// 兼容：code.Error,error
type GrpcError struct {
	prefix  string
	code    int
	mcode   string
	message string
}

// 兼容性检测
var _ code.Error = new(GrpcError)

// Code 返回一个数字错误，兼容code.Error.Code()
func (grpcError *GrpcError) Code() int {
	return grpcError.code
}

// Mcode 返回一个全量的错误码，兼容code.Error.Mcode()
func (grpcError *GrpcError) Mcode() string {
	return grpcError.mcode
}

// Error 返回错误讯息，兼容error.Error
func (grpcError *GrpcError) Error() string {
	return fmt.Sprintf("%s,%s", grpcError.mcode, grpcError.message)
}

// Message 返回错误讯息，兼容error.Error
func (grpcError *GrpcError) Message() string {
	return grpcError.message
}

// CodeError 生成一个错误
func (rsp *CommResponse) CodeError() code.Error {
	if rsp.Result {
		return nil
	}

	return parseCodeError(rsp.Mcode, rsp.Message)
}

// parseCodeError 解析
func parseCodeError(mcode, msg string) *GrpcError {
	if mcode == "" {
		return &GrpcError{
			prefix:  "",
			code:    NotIntCodeError,
			mcode:   "UNKOWN_ERROR",
			message: msg,
		}
	}

	index := strings.LastIndex(mcode, "_")
	if index < 0 || index == len(mcode)-1 {
		return &GrpcError{
			prefix:  mcode,
			code:    NotIntCodeError,
			mcode:   mcode,
			message: msg,
		}
	}

	c, err := strconv.Atoi(mcode[index+1:])
	if err != nil {
		return &GrpcError{
			prefix:  mcode,
			code:    NotIntCodeError,
			mcode:   mcode,
			message: msg,
		}
	}

	return &GrpcError{
		prefix:  mcode[:index],
		code:    c,
		mcode:   mcode,
		message: msg,
	}
}
