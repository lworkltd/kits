package rabbit

import (
	"testing"
)

func TestMakeupSettings(t *testing.T) {
	MakeupSettings(
		NewExchangeSettings().AutoDelete(true).Durable(true).Internal(true),
		NewQueueSettings().AutoDelete(true).Exclusive(false).Durable(false),
		NewConsumeSettings().AutoAck(true).Exclusive(true).NoLocal(true),
	)
}
