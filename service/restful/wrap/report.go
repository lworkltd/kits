package wrap

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/restful/code"
)

var Report func(err code.Error, ctx *gin.Context, status int, since time.Time)
