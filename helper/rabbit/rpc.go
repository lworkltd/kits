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

type RPCUtil struct {
	*Session
	records      map[string]chan *amqp.Delivery
	recordsMutex sync.Mutex
	queueName    string
	timeout      time.Duration
}

func NewRPCUtil(sess *Session, timeout time.Duration) *RPCUtil {
	util := &RPCUtil{
		Session: sess,
		records: make(map[string]chan *amqp.Delivery),
		timeout: timeout,
	}

	return util
}

func (util *RPCUtil) SetupQueue(name string) error {
	handler := func(dli *amqp.Delivery) {
		fmt.Println(dli.CorrelationId)
		util.recordsMutex.Lock()
		record, ok := util.records[dli.CorrelationId]
		util.recordsMutex.Unlock()

		if !ok {
			return
		}

		record <- dli
		close(record)

		util.recordsMutex.Lock()
		delete(util.records, dli.CorrelationId)
		util.recordsMutex.Unlock()
	}

	queue, err := util.DeclareAndHandleQueue(name, handler, map[string]bool{
		"queue/durable":    false,
		"queue/autodelete": true,
		"queue/exclusive":  false,
	})
	if err != nil {
		return err
	}

	util.queueName = queue.Name

	return err
}

func (util *RPCUtil) PublishBytes(b []byte, ex, routingKey string, args map[string]string) (*amqp.Delivery, error) {
	cid := uuid.New().String()
	args["correlationid"] = cid
	args["replyto"] = util.queueName
	args["expiration"] = strconv.Itoa(int(util.timeout / time.Millisecond))
	err := util.Session.PublishBytes(b, ex, routingKey, args)
	if err != nil {
		return nil, err
	}

	waitChan := make(chan *amqp.Delivery, 1)
	util.recordsMutex.Lock()
	util.records[cid] = waitChan
	util.recordsMutex.Unlock()

	select {
	case <-time.After(util.timeout):
		util.recordsMutex.Lock()
		util.records[cid] = waitChan
		util.recordsMutex.Unlock()

		return nil, errors.New("TIMEOUT")
	case delivery := <-waitChan:
		return delivery, nil
	}
}
