package rabbit

import (
	"sync"

	"time"

	"errors"

	"strconv"

	"fmt"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type RpcUtil struct {
	sess *Session
	sync.Mutex

	records map[string]chan *amqp.Delivery
	replyTo string
	timeout time.Duration
}

func NewRpcUtil(sess *Session, timeout time.Duration) *RpcUtil {
	util := &RpcUtil{
		sess:    sess,
		records: make(map[string]chan *amqp.Delivery),
		timeout: timeout,
	}

	return util
}

func (util *RpcUtil) SetupReplyQueue(name string) error {
	handler := func(dli *amqp.Delivery) {
		util.Lock()
		record, ok := util.records[dli.CorrelationId]
		util.Unlock()

		if !ok {
			return
		}

		record <- dli
		close(record)

		util.Lock()
		delete(util.records, dli.CorrelationId)
		util.Unlock()
	}

	queue, err := util.sess.DeclareAndHandleQueue(name, handler,
		NewQueueSettings().AutoDelete(true).Durable(false).Exclusive(false))
	if err != nil {
		return err
	}

	util.replyTo = queue.Name

	return err
}

func (util *RpcUtil) Call(b []byte, ex, queueOrKey string, options ...PublishOption) (*amqp.Delivery, error) {
	if queueOrKey == "" {
		return nil, fmt.Errorf("amqp rpc must specify a routing key or queue name")
	}

	if util.replyTo == "" {
		return nil, fmt.Errorf("amqp rpc need a reply queue")
	}

	cid := uuid.New().String()
	options = append(options, OptionReplyTo(util.replyTo))
	options = append(options, OptionCorrelationId(cid))
	options = append(options, OptionExpiration(strconv.Itoa(int(util.timeout/time.Millisecond))))

	err := util.sess.Publish(b, ex, queueOrKey, options...)
	if err != nil {
		return nil, err
	}

	waitChan := make(chan *amqp.Delivery, 1)
	util.Lock()
	util.records[cid] = waitChan
	util.Unlock()

	select {
	case <-time.After(util.timeout):
		util.Lock()
		util.records[cid] = waitChan
		util.Unlock()
		return nil, errors.New("amqp rpc time out")
	case delivery := <-waitChan:
		return delivery, nil
	}
}
