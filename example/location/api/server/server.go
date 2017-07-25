package server

import (
	"github.com/gin-gonic/gin"
	"github.com/lvhuat/kits/example/location/api/errcode"
	"github.com/lvhuat/kits/example/location/position"
	"github.com/lvhuat/kits/pkgs/ginutil"
	"github.com/lvhuat/kits/service/profile"
	"github.com/lvhuat/kits/service/restful/code"
	"github.com/lvhuat/kits/service/restful/wrap"
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
	r := gin.Default()

	if err := initService(r, option); err != nil {
		return err
	}

	root := r.Group("/")
	if option.PathPrefix != "" {
		root = root.Group(option.PathPrefix)
	}
	v1 := root.Group("/v1")
	wrapper.Get(v1, "/bar", GetCitizenLocation)

	return r.Run(option.Host)
}

func checkIndentifyValid(id string) bool {
	if len(id) != 15 || len(id) != 18 {
		return false
	}
	return true
}

func GetCitizenLocation(ctx *gin.Context) (interface{}, code.Error) {
	id := ctx.Params.ByName("id")
	if id == "" {
		return nil, code.Newf(errcode.LackParameter, "citizen identify required")
	}

	if valid := checkIndentifyValid(id); !valid {
		return nil, code.Newf(errcode.LackParameter, "bad citizen identify %s", id)
	}

	return position.GetCitizenPosition(ginutil.CtxFromGinContext(ctx), id)
}
