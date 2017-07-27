package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/example/citizen/api/server"
	"github.com/lworkltd/kits/example/citizen/conf"
	"github.com/lworkltd/kits/example/citizen/model"
	"gopkg.in/mgo.v2"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	if err := conf.Parse("app.toml"); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to boot service")
	}

	conf.Dump()

	session, err := mgo.Dial(conf.GetMongo().Url)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to dail mongo")
	}
	model.Init(&model.Option{Session: session})

	if err := server.Setup(conf.GetService()); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to boot service")
	}
}
