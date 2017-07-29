package server

import (
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/example/location/api/errcode"
	"github.com/lworkltd/kits/example/location/position"
	"github.com/lworkltd/kits/service/profile"
	"github.com/lworkltd/kits/service/restful/code"
	"github.com/lworkltd/kits/service/restful/wrap"
	discoveryutils "github.com/lworkltd/kits/utils/discovery"
	ginutils "github.com/lworkltd/kits/utils/gin"
)

var wrapper *wrap.Wrapper

// TODO: 在gin所监听的接口同时处理pprof
func initService(_ *gin.Engine, option *profile.Service) error {
	wrapper = wrap.New(&wrap.Option{
		Prefix: option.McodePrefix,
	})

	if option.Reportable {
		// TODO:report service
	}

	if option.PprofEnabled {
		// TODO:handle for pprof
	}

	return discoveryutils.RegisterServerWithProfile("/ping", option)
}

func Setup(option *profile.Service) error {
	r := gin.New()
	r.Use(gin.Recovery())
	if err := initService(r, option); err != nil {
		return err
	}

	root := r.Group("/")
	if option.PathPrefix != "" {
		root = root.Group(option.PathPrefix)
	}
	root.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1 := root.Group("/v1")
	wrapper.Get(v1, "/location", GetCitizenLocation)

	return r.Run(option.Host)
}

func checkIndentifyValid(id string) bool {
	if len(id) != 15 && len(id) != 18 {
		return false
	}
	return true
}

func GetCitizenLocation(ctx *gin.Context) (interface{}, code.Error) {
	id := ctx.Query("id")
	if id == "" {
		return nil, code.Newf(errcode.LackParameter, "citizen identify required")
	}

	if valid := checkIndentifyValid(id); !valid {
		return nil, code.Newf(errcode.LackParameter, "bad citizen identify %s", id)
	}

	return position.GetCitizenPosition(ginutils.CtxFromGinContext(ctx), id)
}
