package discover

import (
	"fmt"

	"github.com/lvhuat/kits/helper/consul"
)

// DiscoverImpl 是服务发现的实现
type DiscoverImpl struct {
	consul *consul.ConsulClient
	static *StaticDiscoverer
}

// Discover 发现服务
func (discover *DiscoverImpl) Discover(service string) ([]string, []string, error) {
	if discover.static != nil {
		remotes, _ := discover.static.Discover(service)
		if len(remotes) != 0 {
			return remotes, remotes, nil
		}
	}

	if discover.consul != nil {
		return discover.consul.Discover(service)
	}

	return nil, nil, fmt.Errorf("not avaliable discover")
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
