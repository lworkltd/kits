package server

import (
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/example/citizen/api/errcode"
	"github.com/lworkltd/kits/example/citizen/model"
	"github.com/lworkltd/kits/example/citizen/person"
	"github.com/lworkltd/kits/service/restful/code"
)

func postPersonInfo(ctx *gin.Context) (interface{}, code.Error) {
	info := &model.PersonInfo{}
	if err := ctx.BindJSON(info); err != nil {
		return nil, code.NewError(errcode.JsonError, err)
	}

	if len(info.Id) != 15 && len(info.Id) != 18 {
		return nil, code.New(errcode.BadParameters, "bad identify")
	}

	if info.Name == "" {
		return nil, code.New(errcode.BadParameters, "name is required")
	}

	return nil, person.AddPersion(ctxFromGinContext(ctx), info)
}
