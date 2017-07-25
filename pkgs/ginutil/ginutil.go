package ginutil

import (
	"context"
	"github.com/gin-gonic/gin"
)

func CtxFromGinContext(ctx *gin.Context) context.Context {
	//TODO:pending the context from gin conetext
	return context.Background()
}
