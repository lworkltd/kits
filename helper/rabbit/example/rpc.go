package main

import (
	"fmt"
	"github.com/lworkltd/kits/helper/rabbit"
	"time"
)

func main() {
	sess, err := rabbit.Dial("amqp://guest:quest@127.0.0.1/test")
	if err != nil {
		panic(err)
	}
	defer sess.Close()

	rpc := rabbit.NewRpcUtil(sess, time.Minute)
	if err := rpc.SetupReplyQueue(""); err != nil {
		panic(err)
	}

	delivery, err := rpc.Call(
		[]byte("hello"),
		"",
		"test-rpc-queue",
		rabbit.OptionReplyTo("test-rpc-reply-queue"),
	)if err != nil {
		panic(err)
	}

	fmt.Println("rpc replied:", string(delivery.Body))
}
