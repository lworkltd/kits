配置约定
------
服务配置

优先原则
------
`ParseAfter`优先，环境变量其次，再是toml，`ParseBefore`最后

固定配置
-------
#### [base]  
```
[base]
go_max_procs=3  # 最大处理线程数量 
mode="dev"  
```

#### [logger]
```
[logger]
format="json"
level="warn"
time_format="RFC3399"
hooks=[
    ["Airbrake","123","xyz","production"],
    ["Syslog","udp","localhost:514","info",""],
    ["SelfDefined","arg1","arg2","arg3"],
    ["FileHook",">=warn|fileline|server.error.log","<=info|server.info.log"]
]
```

#### [service]
```
[service]
host=":8080"
path_prefix="myservice"
mcode_profix="kits_simple_"

trace_enabled=true
access_log_enabled=true

reportable=true
report_ip="${ip_of_interface,}"
report_tags=["master","v0.2.1"]
report_health=true
report_name="simple-service2"
report_id="kits_simple_8080"

handle_pprof = true
handle_pprof_prefix="/abcdefg"
```

#### [redis]
```
[redis]
endpoints="${kv_of_consul,kits/redis/endpoint-external}" 
password="abc123"                                      
db=1                                                   
```

#### [mongo]
```
[mongo]
url="${kv_of_consul,kits/mongo/url-external}"
```

#### [service]
```
[service]
host=":8080"
path_prefix="myservice"
mcode_profix="kits_simple_"

trace_enabled=true
access_log_enabled=true

reportable=true
report_ip="${ip_of_interface,}"
report_tags=["master","v0.2.1"]
report_health=true
report_name="simple-service2"
report_id="kits_simple_8080"

handle_pprof = true
handle_pprof_prefix="/abcdefg"
```

#### [consul]
```
[consul]
endpoint="120.76.46.55:8500"
auto_sync_enabled = true
```

#### [discovery]
```
[discovery]
enable_consul = true
enable_static = true
static_services=[
    "rancher ${kv_of_consul,kits/rancher/url-external}",
    "product_service 127.0.0.1:1923 127.0.0.1:2023"
]
```

#### [hystrix]
```
[hystrix]
statsd_url="${kv_of_consul,kits/statsd/url-external}"
prefix="xyz.123"
timeout=1000
max_concurrent_request=200
error_percent_threshold=20
# service_circuits = [
#     ["serviceA",1000,200,20,
#         ["seviceA-id-1",1000,200,20],
#         ["seviceA-id-2",1000,200,20]
#     ],
#     ["serviceB",-1,-1,-1,
#         ["seviceA-id-1",1000,200,20],
#         ["seviceA-id-2",1000,-1,20]
#     ]
# ]
```

#### [invoker]
```
[invoker]
load_balance_mode = "round-robin"
tracing_enabled = true
hytrix_enabled = true
logger_enabled = true
```


其他说明
----
+ **eval**  
对于有些动态配置，比如consul的kv，网卡的ip等可以使用eval来解释配置的实际值
,但是要使用这种规则首先需要在代码中对实现接口,参考[Eval指引](../pkgs/eval/README.md)
+ **invoker.balance_mode**  
目前支持`round-robin`
+ **hystrix.service_circuits**  
对于不同服务甚至是服务ID的熔断配置，尚未实现
+ **redis.endpoints**   
如果你是集群用半角逗号分割你的地址
