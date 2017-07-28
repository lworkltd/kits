package main

import (
	"fmt"
	"github.com/lworkltd/kits/helper/rabbit"
	"github.com/streadway/amqp"
)

func exchangeHandler(delivery *amqp.Delivery) {
	fmt.Println("Delivery coming:", string(delivery.Body))
}

func main() {
	sess, err := rabbit.Dial("amqp://guest:quest@127.0.0.1/test")
	if err != nil {
		panic(err)
	}
	defer sess.Close()

	if _, err := sess.HandleExchange(
		"queue-to-bind",
		"exchange-name",
		"topic",
		exchangeHandler,
		rabbit.MakeupSettings(
			rabbit.NewExchangeSettings().Durable(true),
			rabbit.NewQueueSettings().Durable(true).Exclusive(false).AutoDelete(true),
			rabbit.NewConsumeSettings().AutoAck(true).Exclusive(false).NoLocal(false),
		), "fruit.*", "vegetables.*"); err != nil {
		panic(err)
	}

	select {
	case <-sess.Closed:
	}
}
