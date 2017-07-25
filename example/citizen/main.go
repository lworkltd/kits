package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/lvhuat/kits/example/citizen/api/server"
	"github.com/lvhuat/kits/example/citizen/conf"
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

	if err := server.Setup(conf.GetService()); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to boot service")
	}
}
