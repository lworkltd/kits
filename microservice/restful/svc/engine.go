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
}

// Service 获取一个服务
func (engine *Engine) Service(name string) IService {
	engine.mutex.RLock()
	service, exsit := engine.serviceMap[name]
	engine.mutex.RUnlock()

	if !exsit {
		service = engine.newService(name, engine.dv)
		engine.serviceMap[name] = service
	}

	return service
}

// SetDiscovery 设置服务发现函数
func (engine *Engine) SetDiscovery(f DiscoveryFunc) {
	engine.dv = f
}

func (engine *Engine) newService(serviceName string, discovery DiscoveryFunc) IService {
	return &ServiceImpl{
		disconvery: discovery,
		name:       serviceName,
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
