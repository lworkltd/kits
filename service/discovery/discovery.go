package discovery

import (
	"fmt"

	"github.com/lworkltd/kits/helper/consul"
)

// DiscoverImpl 是服务发现的实现
type DiscoverImpl struct {
	seacher    func(string) ([]string, []string, error)
	static     func(string) ([]string, []string, error)
	register   func(*consul.RegisterOption) error
	unregister func(*consul.RegisterOption) error
}

// Discover 发现服务
func (discovery *DiscoverImpl) Discover(service string) ([]string, []string, error) {
	if discovery.static != nil {
		remotes, _, _ := discovery.static(service)
		if len(remotes) != 0 {
			return remotes, remotes, nil
		}
	}

	if discovery.seacher != nil {
		return discovery.seacher(service)
	}

	return nil, nil, fmt.Errorf("not avaliable discovery")
}

// Register 注册服务
func (discovery *DiscoverImpl) Register(option *consul.RegisterOption) error {
	if discovery.register == nil {
		return fmt.Errorf("service register not initialize yet")
	}

	return discovery.register(option)
}

// Unregister 删除服务
func (discovery *DiscoverImpl) Unregister(option *consul.RegisterOption) error {
	if discovery.unregister == nil {
		return fmt.Errorf("service unregister not initialize yet")
	}

	return discovery.unregister(option)
}
