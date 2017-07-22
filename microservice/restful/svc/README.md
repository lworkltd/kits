svc
--------
提供简便的访问注册在服务发现上的服务器接口

type Request struct{
    X string;
}


type Discovery interface{
     DiscoveryFunc() func(string)([]string,error) // 用consul实现
     Register(string,addr string,port int16,tags []string,healthCheck string)
}
var discovery ConsulDiscovery

type Response struct{
    Result bool `json:result"`
    Mcode  string `json:"mcode"`
    Message string `json:"message"`
    Body json.RawMessage `json:"data,omitempty"`
} 

var res Response
svc.SetDiscovery(discovery.DiscoveryFunc())
svc.Service({serviceName})
.Post("/v1/tenants/{tenant}/servers/{servers}")
.Header({headers}) // 头参数
.Query({querys}) // 查询参数
.Route({routes}) // 路径参数
.Context({context or nil}) // 上下文
.Json({input}) // json负载
.Run(&res)

util.Decode(res,&{})

