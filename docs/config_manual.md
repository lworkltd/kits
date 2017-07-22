配置约定
------
服务配置

优先原则
------
环境变量优先，toml文件配置其次，代码中的默认值最后

固定配置
-------
#### [base]  
用于配置进程资源相关配置  
```
[base]
go_max_procs=0.5  # 最大处理线程数量
```

#### [redis]
```
[redis]
host="kv_of_consul(kits.sample.redis.host)" # redis的地址
password="abc123"                           # redis的密码
db=1                                        # redis的工作DB
```
#### [mongo]
```
[mongo]
url="amqp://root:abc123@127.0.0.1:27017/admin"  # mongo的地址
```

#### [service]
```
[service]
name="simple-service2"              # 注册服务名称
id="kits_simple_8080"               # 注册服务ID
host=":8080"                        # 服务监听端口
reportable=true                     # 服务是否上报
report_ip="ip_of_interface(eth0)"   # 上报地址
report_ip=8080                      # 上报端口
tags=["master","v0.2.1"]            # 上报tag
error_prefix="kits_simple_"         # 服务返回错误码前缀
report_health=true                  # 是否上报安全检查
pprof_enable = true                 # 服务是否启用pprof
pprof_path_prefix="/abcdefg"        # pprof地址前缀
```

#### [consul]
consul配置，如果使用`*_of_consul`的`eval`则必须配置consul
```
[consul]
consul_host="127.0.0.1:8500" #consul的地址
```


关于eval
----
对于有些动态配置，比如consul的kv，网卡的ip等可以使用eval来解释配置的实际值
,但是要使用这种规则首先需要在代码中对实现接口[Eval](https://github.com/lvhuat/kits/blob/master/pkgs/eval/eval.go#L4)  
`if_of_interface(eth0)`  
网卡的`eth0`的端口  
`kv_of_consul(key)`  
consul的键值