eval
-----
计算一个字符串表达式的值，比如consul的键值，从http请求获取的值,只要你注册了方法

```
type ExecutorFunc func(...string) (string, bool, error)
```

使用
---------
```
// 注册执行器
eval.SingleArgsExecutor("ip_of_interface",ipnet.Ipv4)
eval.SingleArgsExecutor("kv_of_consul",consul.KeyValue)

// 使用执行器
eval.Value("$(ip_of_interface,eth0)")
eval.Value("$(kv_of_consul,common.mongo.url)") 
```

如果你有一个结构体,也许它表达了你的配置文件或者其他，那么你可以对结构体进行扫描，这样一来里面的字符串就会被替换：
```
var myStruct AnyStruct
eval.Complete(&myStruct)
```