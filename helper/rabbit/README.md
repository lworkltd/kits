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
// durable=true,autodelete=false 始终保持
// durable=true,autodelete=true 服务启动时无绑定关系即删除
// durable=false,autodelete=true 服务启动时无绑定关系即删除
// durable=false,autodelete=false 服务启动时删除
"exchange/durable":    true
"exchange/autodelete": false
// 被设置后，不能用于接受消息，但是你可以用它来建立你的内部消息拓扑结构（二级exchange）
"exchange/internal":  false
```
+ queue 
```
// durable=true,autodelete=false 始终保持，仅能绑定durable的exchange
// durable=true,autodelete=true 服务启动时无绑定关系即删除，仅能绑定durable的exchange
// durable=false,autodelete=true 无绑定关系后一会后即删除，仅能绑定非durable的exchange
// durable=false,autodelete=false 服务启动时删除,仅能绑定非durable的exchange
"queue/durabale":   true
"queue/autodelete": false
// 排他队列，同连接可访问，其他连接不可见，名字唯一，连接释放删除(durabale被忽略)
"queue/exclusive":  false 
``
```
+ consume
```
// 在接受到消息以后理解反馈给服务器，注意模块并不理会业务是否处理得过来
"consume/autoack":   true
// 排他消费，绑定时会检查队列是否存在其他消费者，
"consume/exclusive": false
// RabbitMQ不支持此设置
"consume/nolocal":   false
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