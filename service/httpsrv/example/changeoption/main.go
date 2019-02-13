package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/httpsrv"
	"github.com/lworkltd/kits/service/httpsrv/httpstat"
	"github.com/lworkltd/kits/service/restful/code"
	"github.com/sirupsen/logrus"
)

func main() {
	testDemo()

	// 自定义上报的方式
	//testReplaceReportFunc()
	// 自定义写日志的方式
	//testReplaceLogFunc()
	// 自定义结果写IO
	//testReplaceWriteResultFunc()
	// 完全自定义
	//testReplaceWrapFunc()
}

// 标准DEMO
func testDemo() {
	wrapper := httpsrv.New(&httpsrv.Option{
		Prefix: "USER",
	})
	//rand.Seed(time.Now().UnixNano())
	wrapper.HandleStat()
	wrapper.Get("/v1/hello", func(ctx *gin.Context) (interface{}, code.Error) {
		//time.Sleep(time.Duration(rand.Int63()) % time.Minute)
		return []string{"Hello", "World"}, nil
	})

	wrapper.Get("/v1/error", func(ctx *gin.Context) (interface{}, code.Error) {
		return nil, code.NewMcode("ERROR", "test error")
	})

	go func() {
		for {
			time.Sleep(time.Second * 20)
			httpstat.Reset()
		}
	}()

	wrapper.Run(":8080")
}

// 替换封装函数
// 使用者可以根据自己的需要替换最主要的转换函数，但是其中所有的东西都得使用者负责
func testOptionReplaceWrapFunc() {
	wrapper := httpsrv.New(&httpsrv.Option{
		WrapFunc: func(fx interface{}) gin.HandlerFunc { // 替换部分
			f := fx.(func(userId string, ctx *gin.Context) (interface{}, error))
			return func(ctx *gin.Context) {
				userId := ctx.Query("userId")
				data, err := f(userId, ctx)
				if err != nil {
					ctx.AbortWithError(400, err)
					return
				}
				ctx.JSON(200, data)
			}
		},
	})

	wrapper.Get("/v1/hello", func(userId string, ctx *gin.Context) (interface{}, error) {
		return []string{"Hello", "World"}, nil
	})
	wrapper.Get("/v1/error", func(userId string, ctx *gin.Context) (interface{}, error) {
		return nil, code.NewMcodef("ERROR", "error happened")
	})

	v2 := wrapper.Group("/v2")
	v2.Get("/hello", func(userId string, ctx *gin.Context) (interface{}, error) {
		return []string{"Hello", "World"}, nil
	})

	v2.Get("/error", func(userId string, ctx *gin.Context) (interface{}, error) {
		return nil, code.New(1000, "error")
	})

	wrapper.Run(":8080")
}

// 替换上报函数
// 自定义上报方式
func testOptionReplaceReportFunc() {
	wrapper := httpsrv.New(&httpsrv.Option{
		Report: func(err code.Error, httpCtx *gin.Context, status int, beginTime time.Time) { // 替换部分
			fmt.Println(err)
		},
	})
	wrapper.Get("/v1/error", func(ctx *gin.Context) (interface{}, code.Error) {
		return nil, code.NewMcodef("ERROR", "error happened")
	})

	wrapper.Run(":8080")
}

// 替换日志打印函数
// 自定义日志方式和格式
func testOptionReplaceLogFunc() {
	wrapper := httpsrv.New(&httpsrv.Option{
		LogFunc: func(logger *logrus.Logger, ctx *gin.Context, status int, since time.Time, data interface{}, cerr code.Error) { // 替换部分
			logger.Infof("Result %v path=%v cost=%v", cerr != nil, ctx.Request.URL.Path, time.Now().Sub(since))
		},
	})

	wrapper.Get("/v1/error", func(ctx *gin.Context) (interface{}, code.Error) {
		return nil, code.NewMcodef("ERROR", "error happened")
	})
	wrapper.Get("/v1/hello", func(ctx *gin.Context) (interface{}, code.Error) {
		return []string{"Hello", "World"}, nil
	})

	wrapper.Run(":8080")
}

// 替换写结果的函数
func testOptionReplaceWriteResultFunc() {
	wrapper := httpsrv.New(&httpsrv.Option{
		WriteResult: func(ctx *gin.Context, marshal httpsrv.MarshalFunc, status int, prefix string, data interface{}, cerr code.Error) { // 替换部分
			type userResult struct {
				Code string
				Data interface{} `json:",inline,omitempty"`
			}

			if cerr != nil {
				ctx.JSON(400, &userResult{
					Code: cerr.Mcode(),
					Data: data,
				})
				return
			}

			ctx.JSON(200, &userResult{
				Code: "ok",
				Data: data,
			})
		},
	})

	wrapper.Get("/v1/error", func(ctx *gin.Context) (interface{}, code.Error) {
		return nil, code.NewMcodef("ERROR", "error happened")
	})
	wrapper.Get("/v1/errorwithdata", func(ctx *gin.Context) (interface{}, code.Error) {
		return []string{"Hello", "World"}, code.NewMcodef("ERROR", "error happened")
	})
	wrapper.Get("/v1/hello", func(ctx *gin.Context) (interface{}, code.Error) {
		return []string{"Hello", "World"}, nil
	})
	wrapper.Get("/v1/nodata", func(ctx *gin.Context) (interface{}, code.Error) {
		return nil, nil
	})

	wrapper.Run(":8080")
}
