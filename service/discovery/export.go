package discovery

import "github.com/lworkltd/kits/helper/consul"

// Discovery 定义了服务发现的接口
type Discovery interface {
	Discover(service string) ([]string, []string, error)
	Register(option *consul.RegisterOption) error
	Unregister(option *consul.RegisterOption) error
}

var defaultDiscovery Discovery

// Discover 发现一个服务
func Discover(service string) ([]string, []string, error) {
	return defaultDiscovery.Discover(service)
}

// Register 发现服务
func Register(option *consul.RegisterOption) error {
	return defaultDiscovery.Register(option)
}

// Unregister 删除服务注册
func Unregister(option *consul.RegisterOption) error {
	return defaultDiscovery.Unregister(option)
}

// Option 初始化服务发现的
type Option struct {
	// StaticFunc 返回静态服务，静态服务比发现服务更加优先，经常用于配置文件写死得服务配置
	// 输入参数为服务名称，第一个返回参数为ip:port列表，第二个为服务ID名称
	// 如果不填写，那么意味着没有静态服务，模块将尝试从SearchFunc获取服务访问地址
	StaticFunc func(string) ([]string, []string, error)

	// SearchFunc 发现服务多用于动态得服务配置
	// 输入参数为服务名称，第一个返回参数为ip:port列表，第二个为服务ID名称列表
	// 如果StaticFunc和SearchFunc都不设置，那么发现服务时将报错
	SearchFunc func(string) ([]string, []string, error)

	// RegisterFunc 注册服务
	RegisterFunc func(*consul.RegisterOption) error

	// 注销服务，仅需填写`Id`
	UnregisterFunc func(*consul.RegisterOption) error
}

// Init 初始化服务发现
func Init(option *Option) error {
	dis := &DiscoverImpl{}
	dis.static = option.StaticFunc
	dis.seacher = option.SearchFunc
	dis.register = option.RegisterFunc
	dis.unregister = option.UnregisterFunc
	defaultDiscovery = dis

	return nil
}
