package discover

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
func (discover *DiscoverImpl) Discover(service string) ([]string, []string, error) {
	if discover.static != nil {
		remotes, _, _ := discover.static(service)
		if len(remotes) != 0 {
			return remotes, remotes, nil
		}
	}

	if discover.seacher != nil {
		return discover.seacher(service)
	}

	return nil, nil, fmt.Errorf("not avaliable discover")
}

// Register 注册服务
func (discover *DiscoverImpl) Register(option *consul.RegisterOption) error {
	if discover.register == nil {
		return fmt.Errorf("service register not initialize yet")
	}

	return discover.register(option)
}

// Unregister 删除服务
func (discover *DiscoverImpl) Unregister(option *consul.RegisterOption) error {
	if discover.unregister == nil {
		return fmt.Errorf("service unregister not initialize yet")
	}

	return discover.unregister(option)
}
