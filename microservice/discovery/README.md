discovery 包
-----
提供服务发现的接口

功能列表
-----
1.支持静态的服务发现  
2.支持consul的服务发现   
3.支持consul的服务注册   
4.支持自动更新服务信息以提升访问效率，同时也支持将很久不用的服务从自动更新列表里面移除  

使用方法
----
```
csl, err := consul.New("10.25.100.164:8500")
if err != nil {
    panic(err)
}

Init(&Option{
    ConsulClient: csl,
})
key := "kits/unittest/hello"
value, _, e := csl.KeyValue("kits/unittest/hello")
if e != nil || value != "world" {
    log.Errorf("key %s in consul,expect %v,get %s,err=%v", key, "world", value, e)
    return
}

o := &consul.RegisterOption{
    Name: "kits-test-server",
    Id:   "kits-test-server-001",
    Ip:   "localhost",
    Port: 11111,
}

Register(o)
remotes, err := Discover(o.Name)
if err != nil || len(remotes) != 1 {
    log.Errorf("expect 1 server got %v ,err=%v", len(remotes), err)
    return
}
if remotes[0] != fmt.Sprintf("%s:%d", o.Ip, o.Port) {
    log.Errorf("expect localhost:11111 server got %v", remotes[0])
    return
}

Unregister(o)

remotes, err = Discover(o.Name)
if err != nil || len(remotes) != 1 {
    log.Errorf("expect 0 server got %v ,err=%v", len(remotes), err)
    return
}
```