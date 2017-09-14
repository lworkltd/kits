package server

import (
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/example/citizen/api/errcode"
	"github.com/lworkltd/kits/example/citizen/person"
	"github.com/lworkltd/kits/service/restful/code"
	"strconv"
	"github.com/lworkltd/kits/service/context"
)

func getPersonInfo(srvContext context.Context, ctx *gin.Context) (interface{}, code.Error) {
	ageString := ctx.Query("age")
	if ageString == "" {
		return nil, code.New(errcode.LackParameters, "age is required")
	}

	age, err := strconv.Atoi(ageString)
	if err != nil || age <= 0 || age > 160 {
		return nil, code.Newf(errcode.BadParameters, "age out of range,%s", ageString)
	}

	list, cerr := person.GetPerson(ctxFromGinContext(ctx), int8(age))
	if err != nil {
		return nil, cerr
	}

	return list, nil
}
