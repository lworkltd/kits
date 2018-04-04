package invokeimpl

import (
	"fmt"

	"github.com/lworkltd/kits/service/grpcinvoke"
	"github.com/lworkltd/kits/utils/co"
)

// ServiceSelector 负载均衡的服务目标选择器
type ServiceSelector interface {
	Select() (string, string, error)
}

// LbStrategyType 定义负载均衡的类型
type LbStrategyType string

const (
	// RoundRobin 轮询
	RoundRobin LbStrategyType = "round-robin"
)

// DefaultStrategyType 默认为轮询模式
var DefaultStrategyType = RoundRobin

// RoundRobinSelector 服务的负载均衡实现
type RoundRobinSelector struct {
	wellCount      co.Int64
	totalCallCount co.Int64
	discovery      grpcinvoke.DiscoveryFunc
	serviceName    string
}

// NewRoundRobinSelector 创建一个轮询策略的服务器IP选择器
func NewRoundRobinSelector(discovery grpcinvoke.DiscoveryFunc) *RoundRobinSelector {
	return &RoundRobinSelector{
		discovery: discovery,
	}
}

// Select 选择一个服务
func (selector *RoundRobinSelector) Select() (string, string, error) {
	index := selector.totalCallCount.Get()
	defer selector.totalCallCount.Add(1)

	if selector.discovery == nil {
		return "", "", fmt.Errorf("service not found")
	}

	remotes, ids, err := selector.discovery(selector.serviceName)
	if err != nil {
		return "", "", fmt.Errorf("discovery service failed")
	}

	if len(remotes) != len(ids) {
		return "", "", fmt.Errorf("discovery return wrong remotes=%v ids=%v", remotes, ids)
	}

	l := len(remotes)
	if l == 0 {
		return "", "", fmt.Errorf("service not found")
	}
	var use int64
	if l > 1 {
		// 通过轮询策略来访问
		// TODO:支持更多的策略
		use = index % int64(l)
	}
	addr, id := remotes[use], ids[use]

	return addr, id, nil
}

// UniqueAddrSelector 唯一地址选择
type UniqueAddrSelector struct {
	addr string
}

// Select 选择一个服务
func (selector *UniqueAddrSelector) Select() (string, string, error) {
	return selector.addr, selector.addr, nil
}

// NewUniqueAddrSelector 返回一个唯一IP选择器
func NewUniqueAddrSelector(addr string) *UniqueAddrSelector {
	return &UniqueAddrSelector{addr: addr}
}

// NewServiceSelector 根据策略来返回
func NewServiceSelector(strategy LbStrategyType, discovery grpcinvoke.DiscoveryFunc) ServiceSelector {
	switch strategy {
	case RoundRobin:
		return NewRoundRobinSelector(discovery)
	}

	return NewRoundRobinSelector(discovery)
}
