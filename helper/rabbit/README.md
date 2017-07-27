rabbit
------------
封对amqp的实用封装


注意
-----
1. 在使用本工具之前，需要充分的amqp协议知识
2. 包内不处理重连逻辑(也不建议)，请在业务层面做相应处理

关于setting
-----
使用AMQP时设置是很多的，为了灵活并且简洁，使用了map[string]bool的方式来传递设置,以下是默认配置：
+ exchange
```
"exchange/durable":    false
"exchange/autodelete": false
"exchange/exclusive":  false
"exchange/nowait":     false
```
+ queue 
```
"queue/durabale":   true
"queue/autodelete": false
"queue/exclusive":  true
"queue/nowait":     false
```
+ bind
```
"bind/nowait":    false
```
+ autoack
```
"autoack":   false
```
+ consume
```
"consume/autoack":   true
"consume/exclusive": false
"consume/nolocal":   false
"consume/nowait":    false
```

使用方法
-------
### 创建会话，所有的前提
```
sess, err := rabbit.Dail("url-of-amqp")
if err != nil {
    logrus.WithFields(logrus.Fields{
        "error": err,
    }).Error("Dail amqp failed")
    return
}
defer sess.Close()
```
### 监听连接错误
```
<-sess.Closed
```

### 消费制定队列 
```
handle := func(deli *amqp.Delivery) {
    defer deli.Ack(false)
    // TODO:xxx
}
if err = sess.HandleQueue(
    handle,
    "name-of-queue",
    map[string]bool{
        "exchange/durable": true,
        "queue/durable":    true,
        "queue/nowait":     false,
        "queue/exclusive":  false,
        "bind/nowait":      false,
        "consume/durable":  true,
    }, // settings
); err != nil {
    logrus.WithFields(logrus.Fields{
        "error": err,
    }).Error("Handle queue failed")
    return
}
```

### 路由处理
```
routingKeys := []string{
    "food.fruit.*",
    "food.vegetables.*",
}

handle := func(deli *amqp.Delivery) {
    defer deli.Ack(false)
    // TODO:xxx
}

if err = sess.HandleExchange(
    handle,
    map[string]bool{
        "exchange/durable": true,
        "queue/durable":    true,
        "queue/nowait":     false,
        "queue/exclusive":  false,
    }, // settings
    "queue-message-route-to",
    "PORT_DATA",    // exchange
    "topic",        // exchange type
    routingKeys..., // routing key
); err != nil {
    logrus.WithFields(logrus.Fields{
        "error": err,
    }).Error("Handle exchange failed")
    return
}
```

### RPC 调用
```
rpcSess := rabbit.NewRPCUtil(sess, time.Minute)
if err := rpcSess.SetupReplyQueue(""); err != nil {
    panic(err)
}

body, err := proto.Marshal(&pb.Request{})
if err != nil {
    panic(err)
}

deli, err := rpcSess.PublishBytes(
    body,
    "",
    fmt.Sprintf("ANY.%s.*.%s", "filed1", "filed2"),
    map[string]string{
        "content_type": "name-of-request",
    },
)
if err != nil {
    panic(err)
}

rsp := &pb.Response{}
if err := proto.Unmarshal(deli.Body, rsp); err != nil {
    logrus.WithFields(logrus.Fields{
        "error": err,
    }).Error("Bad response proto")
    return
}
// ...
```