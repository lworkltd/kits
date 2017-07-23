package svc

import (
	"errors"
	"sync"
)

// ErrDiscoveryNotConfig 出现在没有设置服务发现函数
var ErrDiscoveryNotConfig = errors.New("discovery not config")

// Engine 提供了向服务发送请求的入口
type Engine struct {
	dv         DiscoveryFunc
	serviceMap map[string]IService
	mutex      sync.RWMutex
	lbMode     string
	useTracing bool
	useCircuit bool
}

func (engine *Engine) Init(option *Option) error {
	engine.dv = option.Discover
	engine.lbMode = option.LoadBalanceMode

	return nil
}

// Service 获取一个服务
func (engine *Engine) Service(name string) IService {
	engine.mutex.RLock()
	service, exsit := engine.serviceMap[name]
	engine.mutex.RUnlock()

	if !exsit {
		service = engine.newService(name, engine.dv)
		engine.mutex.Lock()
		engine.serviceMap[name] = service
		engine.mutex.Unlock()
	}

	return service
}

// Addr 获取一个匿名服务
func (engine *Engine) Addr(addr string) IService {
	return engine.newAddr(addr)
}

// newAddr 创建一个服务
func (engine *Engine) newService(serviceName string, discovery DiscoveryFunc) IService {
	return &ServiceImpl{
		discover:   discovery,
		name:       serviceName,
		useTracing: engine.useTracing,
		useCircuit: engine.useCircuit,
	}
}

// newAddr 创建固定IP的匿名服务
func (engine *Engine) newAddr(addr string) IService {
	discover := func(string) ([]string, error) { return []string{addr}, nil }
	return &ServiceImpl{
		discover: discover,
		name:     addr,
	}
}

func newEngine() *Engine {
	return &Engine{
		dv: func(name string) ([]string, error) {
			return nil, ErrDiscoveryNotConfig
		},
		serviceMap: make(map[string]IService, 10),
	}
}
