package invoke

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/afex/hystrix-go/hystrix"
)

type TestOutput struct {
	Number int
}

func TestExtractHeader(t *testing.T) {
	tests := []struct {
		name       string
		invokeErr  error
		statusCode int
		res        *Response
		out        interface{}
		want       string
		wantNumber int
	}{
		{name: "MyService", invokeErr: nil, statusCode: 200, res: &Response{Result: true}, out: nil, want: ""},
		{name: "MyService", invokeErr: nil, statusCode: 200, res: &Response{Result: true}, out: &TestOutput{}, want: MCODE_INVOKE_FAILED},
		{name: "MyService", invokeErr: nil, statusCode: 200, res: &Response{Result: true, Data: json.RawMessage(`{"Number":12345}`)}, out: &TestOutput{}, want: "", wantNumber: 12345},
		{name: "MyService", invokeErr: nil, statusCode: 200, res: &Response{Result: true, Data: json.RawMessage(`{"Number":"12345"}`)}, out: &TestOutput{}, want: MCODE_INVOKE_FAILED},
		{name: "MyService", invokeErr: nil, statusCode: 200, res: &Response{}, out: nil, want: MCODE_INVOKE_FAILED},
		{name: "MyService", invokeErr: nil, statusCode: 404, res: &Response{Result: true}, out: nil, want: MCODE_INVOKE_FAILED},
		{name: "MyService", invokeErr: &url.Error{Err: &net.OpError{Err: &os.SyscallError{}}}, statusCode: 0, res: nil, out: nil, want: MCODE_INVOKE_FAILED},
		{name: "MyService", invokeErr: &url.Error{Err: &net.OpError{}}, statusCode: 0, res: nil, out: nil, want: MCODE_INVOKE_FAILED},
		{name: "MyService", invokeErr: &url.Error{}, statusCode: 0, res: nil, out: nil, want: MCODE_INVOKE_FAILED},
		{name: "MyService", invokeErr: nil, statusCode: 200, res: &Response{Result: false, Code: "SERVICE_ERROR"}, out: nil, want: "SERVICE_ERROR"},
		{name: "MyService", invokeErr: hystrix.ErrTimeout, statusCode: 0, res: &Response{}, out: nil, want: MCODE_INVOKE_TIMEOUT},
		{name: "MyService", invokeErr: hystrix.ErrCircuitOpen, statusCode: 0, res: &Response{}, out: nil, want: MCODE_INVOKE_FAILED},
		{name: "MyService", invokeErr: hystrix.ErrMaxConcurrency, statusCode: 0, res: &Response{}, out: nil, want: MCODE_INVOKE_FAILED},
		{name: "MyService", invokeErr: fmt.Errorf("other invoke errors"), statusCode: 0, res: &Response{}, out: nil, want: MCODE_INVOKE_FAILED},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractHeader(tt.name, tt.invokeErr, tt.statusCode, tt.res, tt.out)
			if tt.want == "" && got == nil {
				if tt.out != nil && tt.wantNumber != tt.out.(*TestOutput).Number {
					t.Errorf("ExtractHeader() out.Number = %v, wantNumber %v", tt.out.(*TestOutput).Number, tt.wantNumber)
				}

				return
			}

			if !reflect.DeepEqual(got.Mcode(), tt.want) {
				t.Errorf("ExtractHeader() = %v, want %v", got.Mcode(), tt.want)
			}
		})
	}
}

type stringReadCloser struct {
	*strings.Reader
}

func (stringReadCloser *stringReadCloser) Close() error {
	return nil
}

func newStringReadCloser(body string) *stringReadCloser {
	return &stringReadCloser{
		Reader: strings.NewReader(body),
	}
}

type errorReaderCloser struct {
}

func (errorReaderCloser *errorReaderCloser) Close() error {
	return nil
}
func (errorReaderCloser *errorReaderCloser) Read(b []byte) (int, error) {
	return 0, fmt.Errorf("read error")
}

func newErrorReaderCloser() *errorReaderCloser {
	return &errorReaderCloser{}
}

func TestExtractHttpResponse(t *testing.T) {
	tests := []struct {
		name      string
		invokeErr error
		rsp       *http.Response
		out       interface{}
		want      string
	}{
		{invokeErr: nil, rsp: &http.Response{StatusCode: 200, Body: newStringReadCloser(`{"mcode":"SERVICE_ERROR"}`)}, out: nil, want: "SERVICE_ERROR"},
		{invokeErr: nil, rsp: &http.Response{StatusCode: 200, Body: newStringReadCloser(``)}, out: nil, want: MCODE_INVOKE_FAILED},
		{invokeErr: nil, rsp: &http.Response{StatusCode: 200, Body: newStringReadCloser(`xxx`)}, out: nil, want: MCODE_INVOKE_FAILED},
		{invokeErr: nil, rsp: &http.Response{StatusCode: 200, Body: newErrorReaderCloser()}, out: nil, want: MCODE_INVOKE_FAILED},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractHttpResponse(tt.name, tt.invokeErr, tt.rsp, tt.out)
			if tt.want == "" && got == nil {
				return
			}

			if !reflect.DeepEqual(got.Mcode(), tt.want) {
				t.Errorf("ExtractHeader() = %v, want %v", got.Mcode(), tt.want)
			}
		})
	}
}
