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

func (util *RPCUtil) SetupReplyQueue(name string) error {
	handler := func(dli *amqp.Delivery) {
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

// readCorrelationId allow user to pass a `cid` or `correlation_id` or use a generated uuid
func readCorrelationId(args map[string]string) string {
	if args == nil {
		return uuid.New().String()
	}

	cid, exist := args["cid"]
	if exist {
		return cid
	}
	cid, exist = args["correlation_id"]
	if exist {
		return cid
	}

	return uuid.New().String()
}

func (util *RPCUtil) PublishBytes(b []byte, ex, queueOrKey string, args map[string]string) (*amqp.Delivery, error) {
	if queueOrKey == "" {
		return nil, fmt.Errorf("amqp rpc must specify a routing key or queue name")
	}

	if util.queueName == "" {
		return nil, fmt.Errorf("amqp rpc need a reply queue")
	}

	cid := readCorrelationId(args)
	args["cid"] = readCorrelationId(args)
	args["reply_to"] = util.queueName
	args["expiration"] = strconv.Itoa(int(util.timeout / time.Millisecond))
	err := util.Session.PublishBytes(b, ex, queueOrKey, args)
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

		return nil, errors.New("amqp rpc time out")
	case delivery := <-waitChan:
		return delivery, nil
	}
}
