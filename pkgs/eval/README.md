eval
-----
获取一个表达式的值

使用
---------
用于解释配置文件中特殊配置
```
// 注册执行器
eval.SingleArgsExecutor("ip_of_interface",ipnet.Ipv4)
eval.SingleArgsExecutor("kv_of_consul",consul.KeyValue)

// 使用执行器
eval.Value("$(ip_of_interface eth0)")
eval.Value("$(kv_of_consul common.mongo.url)") 
```