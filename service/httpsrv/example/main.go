package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/httpsrv"
	"github.com/lworkltd/kits/service/restful/code"
)

func main() {
	testReplaceWrapFunc()
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
