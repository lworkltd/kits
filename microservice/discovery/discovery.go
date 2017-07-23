package discovery

import (
	"fmt"

	"github.com/lvhuat/kits/helper/consul"
)

// Discoverer 定义了服务发现的接口
type Discoverer interface {
	Discover(service string) ([]string, error)
	Register(option *consul.RegisterOption) error
	Unregister(option *consul.RegisterOption) error
}

var defaultDiscoverer Discoverer

// Discover 发现一个服务
func Discover(service string) ([]string, error) {
	return defaultDiscoverer.Discover(service)
}

// DiscoverImpl 是服务发现的实现
type DiscoverImpl struct {
	consul *consul.ConsulClient
	static *StaticDiscoverer
}

// Discover 发现服务
func (discover *DiscoverImpl) Discover(service string) ([]string, error) {
	if discover.static != nil {
		remotes, _ := discover.static.Discover(service)
		if len(remotes) != 0 {
			return remotes, nil
		}
	}

	if discover.consul != nil {
		return discover.consul.Discover(service)
	}

	return nil, fmt.Errorf("not avaliable discover")
}

// Register 注册服务
func (discover *DiscoverImpl) Register(option *consul.RegisterOption) error {
	if discover.consul == nil {
		return fmt.Errorf("consul not initalize yet")
	}
	return discover.consul.Register(option)
}

// Unregister 删除服务
func (discover *DiscoverImpl) Unregister(option *consul.RegisterOption) error {
	if discover.consul == nil {
		return fmt.Errorf("consul not initalize yet")
	}
	return discover.consul.Unregister(option)
}

// KeyValue 返回一个consul的key值
func (discover *DiscoverImpl) KeyValue(key string) (string, bool, error) {
	if discover.consul == nil {
		return "", false, fmt.Errorf("consul not initalize yet")
	}

	return discover.consul.KeyValue(key)
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
	StaticServices []*StaticService
	ConsulClient   *consul.ConsulClient
}

// InitDisconvery 初始化服务发现
func Init(option *Option) error {
	dis := &DiscoverImpl{}
	dis.static = NewStaticDiscoverer(option.StaticServices)
	dis.consul = option.ConsulClient
	defaultDiscoverer = dis

	return nil
}
