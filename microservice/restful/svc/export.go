package svc

import (
	"context"

	"github.com/golang/protobuf/proto"
)

// DiscoveryFunc 服务发现的函数
type DiscoveryFunc func(name string) ([]string, error)

// Option 用于初始化引擎的参数
type Option struct {
	Discover        DiscoveryFunc
	LoadBalanceMode string
	UseTracing      bool
	UseHystrix      bool
}

// IEngine 引擎
type IEngine interface {
	Service(string) IService // 获取一个服务
	Addr(string) IService    // 创建一个匿名服务
	Init(*Option) error      // 初始化
}

// IService 服务
type IService interface {
	Get(string) Client            // GET
	Post(string) Client           // POST
	Put(string) Client            // PUT
	Delete(string) Client         // DELETE
	Method(string, string) Client // 自定义方法
	Name() string                 //服务名称
}

// Client 客户端
type Client interface {
	Headers(map[string]string) Client    // 添加头部
	Header(string, string) Client        // 添加头部
	Query(string, string) Client         // 添加查询参数
	QueryArray(string, ...string) Client // 添加查询参数
	Querys(map[string][]string) Client   // 添加查询参数
	Route(string, string) Client         // 添加路径参数
	Routes(map[string]string) Client     // 添加路径参数
	Json(interface{}) Client             // 添加Json消息体
	Proto(proto.Message) Client          // 添加Proto对象
	Exec(interface{}) error              // 执行请求
	Context(context.Context) Client      // 上下文
}

var engine IEngine = newEngine()

// 初始化
func Init(option *Option) error {
	return engine.Init(option)
}

// Service 返回服务器
// More examples please see <<README.md>>
func Service(name string) IService {
	return engine.Service(name)
}

func Addr(addr string) IService {
	return engine.Addr(addr)
}
