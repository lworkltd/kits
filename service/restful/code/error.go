package code

import (
	"fmt"
)

type Error interface {
	error
	Code() int // 错误码
	Mcode() string
	Message() string
}

type errorImpl struct {
	message string
	mcode   string
	code    int
}

func (err *errorImpl) Error() string {
	if err.mcode == "" {
		return fmt.Sprintf("%d,%s", err.code, err.message)
	}

	return fmt.Sprintf("%s,%s", err.mcode, err.message)
}

func (err *errorImpl) Message() string {
	return err.message
}

func (err *errorImpl) Code() int {
	return err.code
}

func (err *errorImpl) Mcode() string {
	return err.mcode
}

func New(code int, message string) Error {
	return &errorImpl{
		code:    code,
		message: message,
	}
}

func Newf(code int, format string, args ...interface{}) Error {
	return &errorImpl{
		code:    code,
		message: fmt.Sprintf(format, args...),
	}
}

func Newln(code int, args ...interface{}) Error {
	return &errorImpl{
		code:    code,
		message: fmt.Sprintln(args...),
	}
}

func NewPrefixf(prefix string, code int, format string, args ...interface{}) Error {
	return &errorImpl{
		mcode:   fmt.Sprintf("%s_%d", prefix, code),
		code:    code,
		message: fmt.Sprintf(format, args...),
	}
}

func NewPrefixln(prefix string, code int, args ...interface{}) Error {
	return &errorImpl{
		mcode:   fmt.Sprintf("%s_%d", prefix, code),
		code:    code,
		message: fmt.Sprintln(args...),
	}
}

func NewError(code int, err error) Error {
	return &errorImpl{
		code:    code,
		message: err.Error(), //
	}
}

func NewMcode(mcode string, msg string) Error {
	return &errorImpl{
		mcode:   mcode,
		message: msg,
		code:    -11,
	}
}

func NewMcodef(mcode string, format string, args ...interface{}) Error {
	return &errorImpl{
		mcode:   mcode,
		message: fmt.Sprintf(format, args...),
	}
}
