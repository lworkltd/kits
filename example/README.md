
location
----------
这是一个微服务调用的例子

描述
----------
这里有两个服务，citizen 和 location
+ citizen 接受请求访问公民信息的服务
+ location 接受请求访问公民实时位置(:>)的信息

设计
----------
很明显，公民实时位置是公民信息的一部分，因此`citizen`在接收到请求时会：
1. 从本地数据库(mongo)检索公民基本信息
2. 再从`location`服务发起实时信息调用
3. `location`收到请求后，从缓存(redis)中读取实时信息
4. `citizen`得到响应后再将信息综合返回给客户端

程序组成
----------
+ 服务间调用，需要服务发现(consul)
+ 请求是可追踪的(zipkin)
+ 具有熔断机制(hystrix)
+ 方便运维部署，部分集中化配置(consul),文件(.toml),环境变量,运行时配置(pkgs/eval)
+ http高效框架(gin)
+ 结构化日志(logrus)
+ 存储(mongo)与缓存(redis)
+ 服务间调用易用(service/invoke)

运行
----------
+ 你需要按照需要填写好里面的配置,然后编译第一个服务：

```
cd example/citizen/
go build
./citizen
```

+ 编译并运行好第二个服务
```
cd example/location/
go build
./location
```

+ 添加第一个市民：
```
curl -X POST -d '{"id":"1234567890abcde","name":"神奇四侠","age":23}' localhost:8080/citizen/v1/citizen

// 200
{
    "result": true
}

```

+ 查询这个市民的信息(实际上这个时候，我们并没有在location的缓存中读取到他的位置信息，我们默认返回经纬度都为0的信息)
```
curl localhost:8080/citizen/v1/citizen?age=23
{
    "result": true,
    "data": [
        {
            "id": "1234567890abcde",
            "name": "神奇四侠",
            "age": 23,
            "nation": "",
            "phone": "",
            "longitude": 0,
            "latitude": 0
        },
        {
            "id": "1234567890abcdd",
            "name": "杜小二",
            "age": 23,
            "nation": "",
            "phone": "",
            "longitude": 0,
            "latitude": 0
        }
    ]
}
```
