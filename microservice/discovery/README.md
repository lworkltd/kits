discovery 包
-----
提供服务发现的接口

功能列表
-----
1.支持静态的服务发现  
2.支持consul的服务发现  
3.支持consul的key/value读取  
4.支持consul的服务注册   
5.支持自动更新服务信息以提升访问效率，同时也支持将很久不用的服务从自动更新列表里面移除  

使用方法
----
```
// 初始化
InitDisconvery(&DiscoveryOption{
    ConsulHost: "10.25.100.164:8500",
})

// 键值查询
key := "kits/unittest/hello"
value, e := KeyValue(key)
if e != nil || value != "world" {
    fmt.Errorf("key %s in consul,expect %v,get %s,err=%v", key, "world", value, e)
    return
}

o := &RegisterOption{
    Name: "kits-test-server",
    Id:   "kits-test-server-001",
    Ip:   "localhost",
    Port: 11111,
}
// 注册服务
Register(o)

// 服务发现
remotes, err := Discover(o.Name)
if err != nil || len(remotes) != 1 {
    fmt.Errorf("expect 1 server got %v ,err=%v", len(remotes), err)
}

if remotes[0] != fmt.Sprintf("%s:%d", o.Ip, o.Port) {
    fmt.Errorf("expect localhost:11111 server got %v", remotes[0])
}

// 删除服务注册
Unregister(o)
```