package main

import (
	"fmt"
	"github.com/lworkltd/kits/helper/rabbit"
	"github.com/streadway/amqp"
)

func queueHandler(delivery *amqp.Delivery) {
	fmt.Println("Delivery coming:", string(delivery.Body))
}

func main() {
	sess, err := rabbit.Dial("amqp://guest:quest@127.0.0.1/test")
	if err != nil {
		panic(err)
	}
	defer sess.Close()

	if _, err := sess.DeclareAndHandleQueue(
		"handle-queue", queueHandler,
		rabbit.MakeupSettings(
			rabbit.NewQueueSettings().Durable(true).Exclusive(false).AutoDelete(true),
			rabbit.NewConsumeSettings().AutoAck(true).Exclusive(false).NoLocal(false),
		)); err != nil {
		panic(err)
	}

	select {
	case <-sess.Closed:
	}
}
