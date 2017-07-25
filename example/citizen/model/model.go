package model

import (
	"gopkg.in/mgo.v2"
)

type Option struct {
	Session *mgo.Session
}

var (
	personCollection *PersonCollection
)

func Person() *PersonCollection {
	return personCollection
}

func Init(option *Option) {
	personCollection = newPersonCollection(option.Session, "data", "person")
}
