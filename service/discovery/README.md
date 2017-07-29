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
consulClient, err := consul.New("your-consul-server)
if err != nil {
    panic(err)
}

Init(&Option{
    ConsulClient: consulClient,
})

option := &consul.RegisterOption{
    Name: "kits-test-server",
    Id:   "kits-test-server-001",
    Ip:   "localhost",
    Port: 11111,
}

Register(option)
remotes, err := Discover(option.Name)
if err != nil || len(remotes) != 1 {
    log.Errorf("expect 1 server got %v ,err=%v", len(remotes), err)
    return
}
if remotes[0] != fmt.Sprintf("%s:%d", o.Ip, o.Port) {
    log.Errorf("expect localhost:11111 server got %v", remotes[0])
    return
}

Unregister(option)

remotes, err = Discover(o.Name)
if err != nil || len(remotes) != 1 {
    log.Errorf("expect 0 server got %v ,err=%v", len(remotes), err)
    return
}
```