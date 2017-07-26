package person

import (
	"context"

	"github.com/Sirupsen/logrus"
	"github.com/lvhuat/kits/example/citizen/model"
	loc "github.com/lvhuat/kits/example/location/model"
	"github.com/lvhuat/kits/pkgs/invokeutil"
	"github.com/lvhuat/kits/service/invoke"
	"github.com/lvhuat/kits/service/restful/code"
)

type PersonReatimeInfo struct {
	*model.PersonInfo
	loc.Location
}

func GetPerson(cxt context.Context, age int8) ([]*PersonReatimeInfo, code.Error) {
	persons, cerr := model.Person().GetPersonAgeGte(age)
	if cerr != nil {
		return nil, cerr
	}

	infos := make([]*PersonReatimeInfo, 0, len(persons))
	for _, person := range persons {
		info := &PersonReatimeInfo{
			PersonInfo: person,
		}
		infos = append(infos, info)
		// TODO: for the service invokation should move to the common packge
		// TODO: performace advice,read the location once
		location := loc.Location{}
		res := &invokeutil.Response{}
		info.Location = location
		status, invokeerr := invoke.Name("kits-location").
			Get("/v1/location").
			Query("id", person.Id).
			Exec(res)
		cerr := invokeutil.Unpkg("kits-location", invokeerr, status, res, &location)
		if cerr != nil {
			logrus.WithFields(logrus.Fields{
				"id":    person.Id,
				"error": cerr.Error(),
				"code":  cerr.Mcode(),
			}).Warn("Read person location failed")
			continue
		}
	}

	return infos, nil
}

func AddPersion(cxt context.Context, p *model.PersonInfo) code.Error {
	return model.Person().AddPerson(p)
}
