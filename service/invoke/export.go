package invoke

import (
	"context"
	"net/http"
)

var doLogger = true

// DiscoveryFunc 服务发现的函数
type DiscoveryFunc func(name string) ([]string, []string, error)

// Option 用于初始化引擎的参数
type (
	Option struct {
		Discover                     DiscoveryFunc
		LoadBalanceMode              string
		UseTracing                   bool
		UseCircuit                   bool
		DoLogger                     bool
		DefaultTimeout               int
		DefaultMaxConcurrentRequests int
		DefaultErrorPercentThreshold int
	}

	// IEngine 引擎
	Engine interface {
		Service(string) Service // 获取一个服务
		Addr(string) Service    // 创建一个匿名服务
		Init(*Option) error     // 初始化
	}

	// IService 服务
	Service interface {
		Get(string) Client            // GET
		Post(string) Client           // POST
		Put(string) Client            // PUT
		Delete(string) Client         // DELETE
		Method(string, string) Client // 自定义方法

		Name() string // 服务名称
		UseTracing() bool
		UseCircuit() bool
		Remote() (string, string, error) // 获取一个服务地址和ID
	}

	// Client 客户端
	Client interface {
		Headers(map[string]string) Client                                 // 添加头部
		Header(string, string) Client                                     // 添加头部
		Query(string, string) Client                                      // 添加查询参数
		QueryArray(string, ...string) Client                              // 添加查询参数
		Queries(map[string][]string) Client                               // 添加查询参数
		Route(string, string) Client                                      // 添加路径参数
		Routes(map[string]string) Client                                  // 添加路径参数
		Json(interface{}) Client                                          // 添加Json消息体
		Body([]byte) Client                                               // 添加byte消息体
		Hystrix(timeOutMillisecond, maxConn, thresholdPercent int) Client //添加熔断参数
		Tls() Client                                                      // 使用HTTPS
		Context(context.Context) Client                                   // 上下文
		Fallback(func(error) error) Client                                // 失败触发器
		Exec(interface{}) (int, error)                                    // 执行请求
		Response() (*http.Response, error)                                // 执行请求，返回标准的http.Response
	}
)

var eng Engine = newEngine()

// Init 初始化
func Init(option *Option) error {
	doLogger = option.DoLogger
	if true == option.UseCircuit {
		//未设置时的默认值
		if 0 == option.DefaultTimeout {
			option.DefaultTimeout = 1000
		}
		if 0 == option.DefaultMaxConcurrentRequests {
			option.DefaultMaxConcurrentRequests = 200
		}
		if 0 == option.DefaultErrorPercentThreshold {
			option.DefaultErrorPercentThreshold = 20
		}
		//设置值不合理时调整
		if option.DefaultTimeout < 10 {
			option.DefaultTimeout = 10
		} else if option.DefaultTimeout > 10000 {
			option.DefaultTimeout = 10000
		}
		if option.DefaultMaxConcurrentRequests < 30 {
			option.DefaultMaxConcurrentRequests = 30
		} else if option.DefaultMaxConcurrentRequests > 10000 {
			option.DefaultMaxConcurrentRequests = 10000
		}
		if option.DefaultErrorPercentThreshold < 5 {
			option.DefaultErrorPercentThreshold = 5
		} else if option.DefaultErrorPercentThreshold > 100 {
			option.DefaultErrorPercentThreshold = 100
		}
	}
	return eng.Init(option)
}

// Name 返回服务器实例
func Name(name string) Service {
	return eng.Service(name)
}

// Addr 返回一个临时服务
func Addr(addr string) Service {
	return eng.Addr(addr)
}
