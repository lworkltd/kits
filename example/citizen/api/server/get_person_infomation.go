package server

import (
	"github.com/gin-gonic/gin"
	"github.com/lvhuat/kits/example/citizen/api/errcode"
	"github.com/lvhuat/kits/example/citizen/person"
	"github.com/lvhuat/kits/service/restful/code"
	"strconv"
)

func getPersonInfo(ctx *gin.Context) (interface{}, code.Error) {
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
