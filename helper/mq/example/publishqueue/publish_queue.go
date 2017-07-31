package main

import (
	"github.com/lworkltd/kits/helper/mq"
)

func main() {
	sess, err := mq.Dial("amqp://guest:quest@127.0.0.1/test")
	if err != nil {
		panic(err)
	}
	defer sess.Close()
	if err := sess.PublishString(
		"good day today",
		"",
		"queue-to-publish",
		mq.OptionContentType("application/text"),
		mq.OptionAppId("my-service-id"),
		mq.OptionUserId("keto"),
	); err != nil {
		panic(err)
	}
}
