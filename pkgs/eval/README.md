eval
-----
获取一个表达式的值

比如：  
```
eval.Value(ip_of_interface(eth0)) // 获取网卡的ip
eval.Value(mtu_of_interface(eth0)) // 获取网卡的mtu
eval.Value(hosts_of_services(fw-product)) // 获取服务的hosts
```