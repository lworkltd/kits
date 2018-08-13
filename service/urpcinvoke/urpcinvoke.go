package urpcinvoke

import "time"

type DiscoveryFunc func(string) ([]string, []string, error)

var services = map[string]*Service{}
var serviceDiscoveryFunc DiscoveryFunc
var defaultOption Option

// Addr 根据地址获取服务
func Addr(addr string) *Service {
	return getServiceByAddr(addr)
}

// Name 根据服务名称获取服务
func Name(t string) *Service {
	return getServiceByName(t)
}

// Option 用于初始化引擎的参数
type Option struct {
	Discover                     DiscoveryFunc
	LoadBalanceMode              string
	UseTracing                   bool
	UseCircuit                   bool
	DoLogger                     bool
	DefaultTimeout               time.Duration
	DefaultMaxConcurrentRequests int
	DefaultErrorPercentThreshold int
}

// Init 初始化
func Init(option *Option) {
	if true == option.UseCircuit {
		//未设置时的默认值
		if 0 == option.DefaultTimeout/time.Millisecond {
			option.DefaultTimeout = 1000 * time.Millisecond
		}
		if 0 == option.DefaultMaxConcurrentRequests {
			option.DefaultMaxConcurrentRequests = 200
		}
		if 0 == option.DefaultErrorPercentThreshold {
			option.DefaultErrorPercentThreshold = 20
		}

		//设置值不合理时调整
		if option.DefaultTimeout < 10*time.Millisecond {
			option.DefaultTimeout = 10 * time.Millisecond
		} else if option.DefaultTimeout > 10*time.Second {
			option.DefaultTimeout = 10 * time.Second
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
	defaultOption = *option
}
