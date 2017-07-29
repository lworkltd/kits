package main

import (
	"fmt"
	"github.com/lworkltd/kits/helper/mq"
	"time"
)

func main() {
	sess, err := mq.Dial("amqp://guest:quest@127.0.0.1/test")
	if err != nil {
		panic(err)
	}
	defer sess.Close()

	rpc := mq.NewRpcUtil(sess, time.Minute)
	if err := rpc.SetupReplyQueue(""); err != nil {
		panic(err)
	}

	delivery, err := rpc.Call(
		[]byte("hello"),
		"",
		"test-rpc-queue",
		mq.OptionReplyTo("test-rpc-reply-queue"),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("rpc replied:", string(delivery.Body))
}
