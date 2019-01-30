package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/httpsrv"
	"github.com/lworkltd/kits/service/restful/code"
)

func main() {
	//testWrapperDataCodeerror()
	//testWrapperDataError()
	//testWrapperNormal()
	//testwrapperError()
	//testWrapperNoWrapperStatus()
	testWrapperNoWrapper()
}

// 我们公司的默认版本#1
func testWrapperDataCodeerror() {
	wrapper := httpsrv.New(nil)
	wrapper.Get("/v1/hello", func(ginCtx *gin.Context) (interface{}, code.Error) {
		return map[string]interface{}{
			"Hello": "World",
		}, nil
	})
	// > curl 127.0.0.1:8080/v1/hello
	// {"data":{"Hello":"World"},"result":true,"timestamp":1548819012353}

	wrapper.Get("/v1/error", func(ginCtx *gin.Context) (interface{}, code.Error) {
		return map[string]interface{}{
			"Hello": "World",
		}, code.NewMcodef("ERROR", "error message")
	})
	// > curl 127.0.0.1:8080/v1/error
	// {"mcode":"ERROR","message":"error message","result":false,"timestamp":1548819074091

	wrapper.Run(":8080")
}

// 我们公司的默认版本#2
func testWrapperDataError() {
	wrapper := httpsrv.New(nil)
	wrapper.Get("/v1/hello", func(ginCtx *gin.Context) (interface{}, error) {
		return map[string]interface{}{
			"Hello": "World",
		}, nil
	})
	// > curl 127.0.0.1:8080/v1/hello
	// {"data":{"Hello":"World"},"result":true,"timestamp":1548819012353}

	wrapper.Get("/v1/error", func(ginCtx *gin.Context) (interface{}, error) {
		return map[string]interface{}{
			"Hello": "World",
		}, code.NewMcodef("ERROR", "error message")
	})
	// > curl 127.0.0.1:8080/v1/error
	// {"mcode":"ERROR","message":"error message","result":false,"timestamp":1548819074091

	wrapper.Run(":8080")
}

// 原生GIN版本
func testWrapperNormal() {
	wrapper := httpsrv.New(nil)
	wrapper.Get("/v1/hello", func(ginCtx *gin.Context) {
		ginCtx.JSON(200, map[string]interface{}{
			"Hello": "World",
		})
	})
	// > curl 127.0.0.1:8080/v1/hello
	// {"Hello":"World"}

	wrapper.Get("/v1/error", func(ginCtx *gin.Context) {
		ginCtx.JSON(400, map[string]interface{}{
			"errorcode": "ERROR",
			"message":   "error message",
		})
	})
	// > curl 127.0.0.1:8080/v1/error -I
	// HTTP/1.1 400 Bad Request
	// ...
	// {"errorcode":"ERROR","message":"error message"}

	wrapper.Run(":8080")
}

// 我们公司的默认版本#3,不需要返回数据
func testwrapperError() {
	wrapper := httpsrv.New(nil)
	wrapper.Get("/v1/hello", func(ginCtx *gin.Context) error {
		return nil
	})
	// > curl 127.0.0.1:8080/v1/hello
	// {"result":true,"timestamp":1548819737557}

	wrapper.Get("/v1/error", func(ginCtx *gin.Context) error {
		return code.NewMcodef("ERROR", "error message")
	})
	// > curl 127.0.0.1:8080/v1/error
	// {"mcode":"ERROR","message":"error message","result":false,"timestamp":1548819722727}

	wrapper.Run(":8080")
}

// 我们公司的默认版本#4,不需要返回数据
func testwrapperCodeError() {
	wrapper := httpsrv.New(nil)
	wrapper.Get("/v1/hello", func(ginCtx *gin.Context) code.Error {
		return nil
	})
	// > curl 127.0.0.1:8080/v1/hello
	// {"result":true,"timestamp":1548819737557}

	wrapper.Get("/v1/error", func(ginCtx *gin.Context) code.Error {
		return code.NewMcodef("ERROR", "error message")
	})
	// > curl 127.0.0.1:8080/v1/error
	// {"mcode":"ERROR","message":"error message","result":false,"timestamp":1548819722727}

	wrapper.Run(":8080")
}

type noWrapResponse struct {
	Hello string
}

func (noWrapResponse *noWrapResponse) NoWrapperPlease() {}

// 返回数据不需要额外封装
func testWrapperNoWrapperStatus() {
	wrapper := httpsrv.New(nil)
	wrapper.Get("/v1/hello", func(ginCtx *gin.Context) (httpsrv.NoWrapperResponse, int) {
		return &noWrapResponse{
			Hello: "World",
		}, 400
	})
	// > curl 127.0.0.1:8080/v1/hello -i
	// HTTP/1.1 400 Bad Request
	// ...
	// {"Hello":"World"}
	wrapper.Run(":8080")
}

// 返回数据不需要额外封装
func testWrapperNoWrapper() {
	wrapper := httpsrv.New(nil)
	wrapper.Get("/v1/hello", func(ginCtx *gin.Context) httpsrv.NoWrapperResponse {
		return &noWrapResponse{
			Hello: "World",
		}
	})
	// > curl 127.0.0.1:8080/v1/hello
	// {"Hello":"World"}
	wrapper.Run(":8080")
}
