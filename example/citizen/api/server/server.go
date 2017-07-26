package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/profile"
	"github.com/lworkltd/kits/service/restful/wrap"
)

var wrapper *wrap.Wrapper

//TODO: 在gin所监听的接口同时处理pprof
func initService(_ *gin.Engine, option *profile.Service) error {
	wrapper = wrap.New(&wrap.Option{
		Prefix: option.McodeProfix,
	})

	if option.Reportable {
		// TODO:report service
	}

	if option.PprofEnabled {
		// TODO:handle for pprof
	}

	return nil
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
	v1 := root.Group("/v1")
	wrapper.Get(v1, "/citizen", getPersonInfo)
	wrapper.Post(v1, "/citizen", postPersonInfo)

	return r.Run(option.Host)
}

func ctxFromGinContext(ctx *gin.Context) context.Context {
	//TODO:pending the context from gin conetext
	return context.Background()
}
