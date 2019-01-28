package httpsrv

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/restful/code"
	"github.com/sirupsen/logrus"
)

func TestDemo(t *testing.T) {
	// 标准情况下的使用方法
	root := gin.New()
	wrapper := New(&Option{
		SnowSlideLimit: 10,
		Prefix:         "USER",
	})
	wrapper.Get(root, "/v1/hello", func(ctx *gin.Context) (interface{}, code.Error) {
		return []string{"Hello", "World"}, nil
	})
	wrapper.Get(root, "/v1/error", func(ctx *gin.Context) (interface{}, code.Error) {
		return nil, code.NewMcodef("ERROR", "error happened")
	})
	v2 := root.Group("/v2")
	wrapper.Get(v2, "/v1/hello", func(ctx *gin.Context) (interface{}, code.Error) {
		return []string{"Hello", "World"}, nil
	})
	wrapper.Get(v2, "/v1/error", func(ctx *gin.Context) (interface{}, code.Error) {
		return nil, code.New(1000, "error")
	})
}

type userDefaultSnowSlide struct {
}

func (slide *userDefaultSnowSlide) Check(ctx *gin.Context) code.Error {
	return nil
}

func TestDemoReplaceSnowSlide(t *testing.T) {
	// 替换自定义的防雪崩对象
	root := gin.New()
	wrapper := New(&Option{
		Prefix:    "USER",
		SnowSlide: &userDefaultSnowSlide{},
	})
	wrapper.Get(root, "/v1/hello", func(ctx *gin.Context) (interface{}, code.Error) {
		return []string{"Hello", "World"}, nil
	})
}

func TestDemoReplaceLogger(t *testing.T) {
	// 替换自定义的日志打印对象
	root := gin.New()
	wrapper := New(&Option{
		SnowSlideLimit: 10,
		Prefix:         "USER",
		Logger:         logrus.New(),
	})
	wrapper.Get(root, "/v1/hello", func(ctx *gin.Context) (interface{}, code.Error) {
		return []string{"Hello", "World"}, nil
	})
}

func TestReplaceWrapFunc(t *testing.T) {
	// 替换封装函数
	// 使用者可以根据自己的需要替换最主要的转换函数，但是其中所有的东西都得使用者负责
	root := gin.New()
	wrapper := New(&Option{
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
	wrapper.Get(v2, "/v1/hello", func(userId string, ctx *gin.Context) (interface{}, error) {
		return []string{"Hello", "World"}, nil
	})
	wrapper.Get(v2, "/v1/error", func(userId string, ctx *gin.Context) (interface{}, error) {
		return nil, code.New(1000, "error")
	})
}
