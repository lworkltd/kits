package code

import (
	"fmt"
)

type Error interface {
	error
	Code() string    // 错误码
	Message() string // 错误信息
}

type errorImpl struct {
	message string
	code    string
	err     string
}

func (err *errorImpl) Error() string {
	if err.err == "" {
		err.err = fmt.Sprintf("%s,%s", err.code, err.message)
	}

	return err.err
}

func (err *errorImpl) Code() string {
	return err.code
}

func (err *errorImpl) Message() string {
	return err.message
}

func New(code, message string) Error {
	return &errorImpl{
		code:    code,
		message: message,
	}
}
func Newf(code, format string, args ...interface{}) Error {
	return &errorImpl{
		code:    code,
		message: fmt.Sprintf(format, args...),
	}
}

func Newln(code string, args ...interface{}) Error {
	return &errorImpl{
		code:    code,
		message: fmt.Sprintln(args...),
	}
}

func NewPrefixf(prefix, code, format string, args ...interface{}) Error {
	return &errorImpl{
		code:    code,
		message: fmt.Sprintf(format, args...),
	}
}

func NewPrefixln(prefix, code string, args ...interface{}) Error {
	return &errorImpl{
		code:    code,
		message: fmt.Sprintln(args...),
	}
}

func NewError(code string, err error) Error {
	return &errorImpl{
		code:    code,
		message: err.Error(), //
	}
}
