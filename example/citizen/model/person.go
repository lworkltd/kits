package model

import (
	"github.com/lvhuat/kits/example/citizen/api/errcode"
	"github.com/lvhuat/kits/service/restful/code"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type PersonInfo struct {
	Id     string `bson:"_id" json:"id"`
	Name   string `bson:"name" json:"name"`
	Age    int8   `bson:"age" json:"age"`
	Nation string `bson:"nation" json:"nation"`
	Phone  string `bson:"phone_number" json:"phone"`
}

type PersonCollection struct {
	s *mgo.Session
	c string
	d string
}

func (person *PersonCollection) collection(sess *mgo.Session) *mgo.Collection {
	return sess.DB(person.d).C(person.c)
}

func newPersonCollection(s *mgo.Session, db, coll string) *PersonCollection {
	return &PersonCollection{s: s, d: db, c: coll}
}

func (person *PersonCollection) readSession() *mgo.Session {
	return person.s.Clone()
}

func bsonReadPersonAgeGte(age int8) bson.M {
	return bson.M{"age": bson.M{"$gte": age}}
}

func (person *PersonCollection) GetPersonAgeGte(age int8) ([]*PersonInfo, code.Error) {
	sess := person.readSession()
	defer sess.Close()

	coll := person.collection(sess)
	query := coll.Find(bsonReadPersonAgeGte(age))

	var list []*PersonInfo
	if err := query.All(&list); err != nil {
		if err == mgo.ErrNotFound {
			return nil, nil
		}
		return nil, code.NewError(errcode.DatabaseFaild, err)
	}

	return list, nil
}

func (person *PersonCollection) AddPerson(p *PersonInfo) code.Error {
	sess := person.readSession()
	defer sess.Close()

	coll := person.collection(sess)

	if err := coll.Insert(p); err != nil {
		return code.NewError(errcode.DatabaseFaild, err)
	}

	return nil
}
