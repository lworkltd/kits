package discover

import "github.com/lworkltd/kits/helper/consul"

// Discoverer 定义了服务发现的接口
type Discoverer interface {
	Discover(service string) ([]string, []string, error)
	Register(option *consul.RegisterOption) error
	Unregister(option *consul.RegisterOption) error
}

var defaultDiscoverer Discoverer

// Discover 发现一个服务
func Discover(service string) ([]string, []string, error) {
	return defaultDiscoverer.Discover(service)
}

// Register 发现服务
func Register(option *consul.RegisterOption) error {
	return defaultDiscoverer.Register(option)
}

// Unregister 删除服务注册
func Unregister(option *consul.RegisterOption) error {
	return defaultDiscoverer.Unregister(option)
}

// DiscoveryOption 初始化服务发现的
type Option struct {
	StaticFunc     func(string) ([]string, []string, error) // 静态服务
	SearchFunc     func(string) ([]string, []string, error) // 发现服务
	RegisterFunc   func(*consul.RegisterOption) error       // 注册服务
	UnregisterFunc func(*consul.RegisterOption) error       // 注销服务
}

// InitDisconvery 初始化服务发现
func Init(option *Option) error {
	dis := &DiscoverImpl{}
	dis.static = option.StaticFunc
	dis.seacher = option.SearchFunc
	dis.register = option.RegisterFunc
	dis.unregister = option.UnregisterFunc
	defaultDiscoverer = dis

	return nil
}
