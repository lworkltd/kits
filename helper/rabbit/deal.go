package rabbit

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
)

type Session struct {
	url         string           // dial connect of mq server
	conn        *amqp.Connection // connection of amqp
	sendChannel *amqp.Channel    // request channel
	recvChannel *amqp.Channel    // response channel
	keepAlive   bool

	closes chan *amqp.Error
	blocks chan amqp.Blocking
	wait   sync.WaitGroup
	Closed chan bool
}

func Dail(url string) (*Session, error) {
	sess, err := newSession(url)
	if err != nil {
		return sess, err
	}
	return sess, nil
}

func newSession(url string) (*Session, error) {
	sess := &Session{}
	if len(url) != 0 {
		sess.url = url
	}

	// dial mq server
	var err error
	if sess.conn, err = amqp.Dial(sess.url); err != nil {
		return nil, fmt.Errorf("sess dial rabbitMQ failed:%s with:%s", err.Error(), url)
	}

	if err := sess.initSend(); err != nil {
		sess.conn.Close()
		return nil, err
	}

	if err := sess.initRecv(); err != nil {
		sess.conn.Close()
		return nil, err
	}

	sess.closes = make(chan *amqp.Error)
	sess.blocks = make(chan amqp.Blocking)
	sess.conn.NotifyClose(sess.closes)
	sess.conn.NotifyBlocked(sess.blocks)
	sess.keepAlive = true
	sess.Closed = make(chan bool, 1)
	go func() {
		log.Debugln("Watch amqp for close events")
	WATCH_LOOP:
		for {
			select {
			case err := <-sess.closes:
				if err != nil {
					log.Debugln("AMQP close", err)
				} else {
					log.Debugln("AMQP close manually")
				}
				break WATCH_LOOP
			case block := <-sess.blocks:
				log.Debugln("AMQP block ", block)
				continue
			}
		}

		sess.wait.Wait()
		sess.Closed <- true

		close(sess.Closed)
	}()

	return sess, nil
}

func (sess *Session) Close() error {
	sess.keepAlive = false
	sess.conn.Close()
	sess.wait.Wait()
	return nil
}

func (sess *Session) initSend() error {
	var err error
	// request mq channel
	if sess.sendChannel, err = sess.conn.Channel(); err != nil {
		return err
	}

	return nil
}

func (sess *Session) initRecv() error {
	var err error
	// response mq channel
	if sess.recvChannel, err = sess.conn.Channel(); err != nil {
		return err
	}

	return nil
}

func (sess *Session) KeepAlive(keep bool) {
	sess.keepAlive = keep
}

func (sess *Session) Qos(prefetchSize, prefectBytes int, global bool) error {
	return sess.recvChannel.Qos(prefetchSize, prefectBytes, global)
}

func (sess *Session) DeclareQueue(name string, cs map[string]bool) (*amqp.Queue, error) {
	configs := map[string]bool{
		"durabale":   true,
		"autodelete": false,
		"exclusive":  false,
		"nowait":     false,
	}

	filterBooleanConfigs(&configs, "queue/", cs, false)
	queue, err := sess.recvChannel.QueueDeclare(
		name,                  // name
		configs["durabale"],   // durable
		configs["autodelete"], // autoDelete
		configs["exclusive"],  // exclusive
		configs["nowait"],     // nowait
		nil,
	)

	if err != nil {
		return nil, err
	}

	return &queue, nil
}

func (sess *Session) ConsumeQueue(fn func(*amqp.Delivery), queue string, cs map[string]bool) error {
	configs := map[string]bool{
		"autoack":   true,
		"exclusive": false,
		"nolocal":   false,
		"nowait":    false,
	}

	filterBooleanConfigs(&configs, "consume/", cs, false)
	messages, err := sess.recvChannel.Consume(
		queue,
		"",
		configs["autoack"],
		configs["exclusive"],
		configs["nolocal"],
		configs["nowait"],
		nil,
	)
	if err != nil {
		return err
	}
	autoAck := configs["autoack"]
	sess.wait.Add(1)
	acount := 0
	go func() {
		log.WithField("queue", queue).Debugln("Consume start")
		defer sess.wait.Done()
		for msg := range messages {
			if !autoAck {
				msg.Ack(false)
			}
			acount++
			fn(&msg)
		}

		log.WithFields(log.Fields{
			"queue":                 queue,
			"handled_message_count": acount,
		}).Debugln("Handle queue close")
	}()

	return nil
}

func (sess *Session) DeclareAndHandleQueue(queue string, fn func(*amqp.Delivery), configs map[string]bool) (*amqp.Queue, error) {
	queueInfo, err := sess.DeclareQueue(queue, configs)
	if err != nil {
		return nil, err
	}

	if err := sess.ConsumeQueue(fn, queueInfo.Name, configs); err != nil {
		return nil, err
	}

	return queueInfo, nil
}

func (sess *Session) DelareExchange(name, kind string, cs map[string]bool) error {
	configs := map[string]bool{
		"durable":    true,
		"autodelete": false,
		"internal":   false,
		"nowait":     false,
	}

	filterBooleanConfigs(&configs, "exchange/", cs, false)
	log.Debugf("delare exchange(%s,%s):%#v", name, kind, configs)
	return sess.recvChannel.ExchangeDeclare(
		name, kind,
		configs["durable"],
		configs["autodelete"],
		configs["internal"],
		configs["nowait"],
		nil,
	)
}

func (sess *Session) HandleExchange(fn func(*amqp.Delivery), configs map[string]bool, bindQueue, exchange, kind string, routeKeys ...string) error {
	if err := sess.DelareExchange(exchange, kind, configs); err != nil {
		return err
	}

	queueInfo, err := sess.DeclareQueue(bindQueue, configs)
	if err != nil {
		return err
	}

	// TODO: add new setting bind/unbind with default false
	if err := sess.BindKeys(queueInfo.Name, exchange, configs, routeKeys...); err != nil {
		return err
	}

	if err := sess.ConsumeQueue(fn, queueInfo.Name, configs); err != nil {
		return err
	}

	return nil
}

func (sess *Session) UnbindKeys(bindQueue, exchange string, keys ...string) error {
	if exchange == "" || bindQueue == "" || len(keys) == 0 {
		return fmt.Errorf("unbind parameters cannot be empty")
	}

	var retErr error
	for _, key := range keys {
		if err := sess.recvChannel.QueueUnbind(bindQueue, key, exchange, nil); err != nil {
			if retErr == nil {
				retErr = err
			}
		}
	}

	return retErr
}

func (sess *Session) BindKeys(bindQueue, exchange string, cs map[string]bool, keys ...string) error {
	if exchange == "" || bindQueue == "" || len(keys) == 0 {
		return fmt.Errorf("unbind parameters cannot be empty")
	}

	configs := map[string]bool{
		"nowait": false,
	}

	filterBooleanConfigs(&configs, "bind/", cs, false)

	var retErr error
	for _, key := range keys {
		if err := sess.recvChannel.QueueBind(
			bindQueue, key, exchange, configs["nowait"], nil,
		); err != nil {
			if retErr == nil {
				retErr = err
			}
		}
	}

	return retErr
}

func (sess *Session) PushRequest(req proto.Message, reqQueue string, sn string, contentType string, rspQueue string) error {
	body, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	return sess.PublishBytes(
		body,
		"",
		reqQueue,
		map[string]string{
			"contenttype":   contentType,
			"replyto":       rspQueue,
			"correlationid": sn,
			"expiration":    strconv.Itoa(int(time.Minute * 3 / time.Millisecond)),
		},
	)
}

func (sess *Session) PublishBytes(body []byte, exchange, queueOrKey string, args map[string]string) error {
	log.WithFields(log.Fields{
		"send_to":        queueOrKey,
		"exchange":       exchange,
		"content_type":   args["content_type"],
		"reply_to":       args["reply_to"],
		"timestamp":      time.Now().UTC(),
		"correlation_id": args["cid"],
		"expiration":     args["expiration"],
	}).Debug("Publish amqp content")
	if err := sess.sendChannel.Publish(
		exchange,
		queueOrKey,
		false, false,
		amqp.Publishing{
			ContentType:   args["content_type"],
			Body:          body,
			ReplyTo:       args["reply_to"],
			Timestamp:     time.Now().UTC(),
			CorrelationId: args["cid"],
			Expiration:    args["expiration"],
		},
	); err != nil {
		return err
	}

	return nil
}

func (sess *Session) PublishString(msg string, exchange, queueOrKey string, args map[string]string) error {
	return sess.PublishBytes([]byte(msg), exchange, queueOrKey, args)
}
