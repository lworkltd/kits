package main

import (
	"github.com/lworkltd/kits/helper/rabbit"
)

func main() {
	sess, err := rabbit.Dial("amqp://guest:quest@127.0.0.1/test")
	if err != nil {
		panic(err)
	}
	defer sess.Close()
	if err := sess.PublishString(
		"good day today",
		"exchange-to-publish",
		"rontiny-key",
		rabbit.OptionContentType("application/text"),
		rabbit.OptionAppId("my-service-id"),
		rabbit.OptionUserId("keto"),
	); err != nil {
		panic(err)
	}
}
