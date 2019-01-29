package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/httpsrv"
	"github.com/lworkltd/kits/service/restful/code"
	"github.com/sirupsen/logrus"
)

func main() {
	// 自定义上报的方式
	//testReplaceReportFunc()
	// 自定义写日志的方式
	//testReplaceLogFunc()
	// 自定义结果写IO
	//testReplaceWriteResultFunc()
	// 完全自定义
	//testReplaceWrapFunc()
}

func testDemo() {
	// 替换自定义的防雪崩对象
	root := gin.New()
	wrapper := httpsrv.New(&httpsrv.Option{
		Prefix: "USER",
	})
	wrapper.Get(root, "/v1/hello", func(ctx *gin.Context) (interface{}, code.Error) {
		return []string{"Hello", "World"}, nil
	})
	wrapper.Get(root, "/v1/error", func(ctx *gin.Context) (interface{}, code.Error) {
		return nil, code.NewMcode("ERROR", "test error")
	})

	root.Run(":8080")
}

func testReplaceWrapFunc() {
	// 替换封装函数
	// 使用者可以根据自己的需要替换最主要的转换函数，但是其中所有的东西都得使用者负责
	root := gin.New()
	wrapper := httpsrv.New(&httpsrv.Option{
		WrapFunc: func(fx interface{}) gin.HandlerFunc {
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
	wrapper.Get(root, "/v1/hello", func(userId string, ctx *gin.Context) (interface{}, error) {
		return []string{"Hello", "World"}, nil
	})
	wrapper.Get(root, "/v1/error", func(userId string, ctx *gin.Context) (interface{}, error) {
		return nil, code.NewMcodef("ERROR", "error happened")
	})

	v2 := root.Group("/v2")
	wrapper.Get(v2, "/hello", func(userId string, ctx *gin.Context) (interface{}, error) {
		return []string{"Hello", "World"}, nil
	})
	wrapper.Get(v2, "/error", func(userId string, ctx *gin.Context) (interface{}, error) {
		return nil, code.New(1000, "error")
	})

	root.Run(":8080")
}

func testReplaceReportFunc() {
	// 替换上报函数
	root := gin.New()
	wrapper := httpsrv.New(&httpsrv.Option{
		Report: func(err code.Error, httpCtx *gin.Context, beginTime time.Time) {
			fmt.Println(err)
		},
	})
	wrapper.Get(root, "/v1/error", func(ctx *gin.Context) (interface{}, code.Error) {
		return nil, code.NewMcodef("ERROR", "error happened")
	})

	root.Run(":8080")
}

func testReplaceLogFunc() {
	// 替换日志打印函数
	root := gin.New()
	wrapper := httpsrv.New(&httpsrv.Option{
		LogFunc: func(logger *logrus.Logger, ctx *gin.Context, since time.Time, data interface{}, cerr code.Error) {
			logger.Infof("Result %v path=%v cost=%v", cerr != nil, ctx.Request.URL.Path, time.Now().Sub(since))
		},
	})

	wrapper.Get(root, "/v1/error", func(ctx *gin.Context) (interface{}, code.Error) {
		return nil, code.NewMcodef("ERROR", "error happened")
	})
	wrapper.Get(root, "/v1/hello", func(ctx *gin.Context) (interface{}, code.Error) {
		return []string{"Hello", "World"}, nil
	})

	root.Run(":8080")
}

func testReplaceWriteResultFunc() {
	// 替换写结果的函数
	root := gin.New()
	wrapper := httpsrv.New(&httpsrv.Option{
		WriteResult: func(ctx *gin.Context, prefix string, data interface{}, cerr code.Error) {
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

	wrapper.Get(root, "/v1/error", func(ctx *gin.Context) (interface{}, code.Error) {
		return nil, code.NewMcodef("ERROR", "error happened")
	})
	wrapper.Get(root, "/v1/errorwithdata", func(ctx *gin.Context) (interface{}, code.Error) {
		return []string{"Hello", "World"}, code.NewMcodef("ERROR", "error happened")
	})
	wrapper.Get(root, "/v1/hello", func(ctx *gin.Context) (interface{}, code.Error) {
		return []string{"Hello", "World"}, nil
	})
	wrapper.Get(root, "/v1/nodata", func(ctx *gin.Context) (interface{}, code.Error) {
		return nil, nil
	})

	root.Run(":8080")
}
