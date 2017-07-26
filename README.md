kits
-----
kits 是一个公共库，包含了常见的框架组件，易用工具

安装
-----
根据需要安装自己需要的模块
```
git get github.com/lworkltd/kits
```

进度
----
[alpha]

1. 尚未实现针对每个服务的熔断配置
2. 需要实现能够实现调用链路追踪的[LoggerFormater](./pkgs/logutil/json_formatter.go)，此外，还需要更多的运行时信息加入日志
3. 很多代码的测试覆盖率依然比较低
4. 工具目前比较少
5. 文档不完善
6. 前期不小心提交了几个exe，导致工程较大，因此需要彻底删除
7. 没有进行拼写检查，可能有很多的拼写错误
8. 目前调用追踪和熔断都没有加入例子中
9. 日志配置中的hooks没有实现

运行
-----
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
