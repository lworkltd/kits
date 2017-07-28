rabbit
------------
这是对[amqp](github.com/streadway/amqp)库的简单封装

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
sess, err := rabbit.Dial("amqp://guest:quest@127.0.0.1/test")
if err != nil {
    panic(err)
}
defer sess.Close()
```
### 监听连接错误
```
<-sess.Closed
```

### 消费制定队列 
```
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
```

### 路由处理
```
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
```

### RPC 调用
```
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
```

### 推送
```
sess, err := rabbit.Dial("amqp://guest:quest@127.0.0.1/test")
if err != nil {
    panic(err)
}
defer sess.Close()
if err := sess.PublishString(
    "good day today",
    "exchange-to-publish",
    "rontiny-key",
    rabbit.OptionContentType("application/text"),
    rabbit.OptionAppId("my-service-id"),
    rabbit.OptionUserId("keto"),
); err != nil {
    panic(err)
}
```