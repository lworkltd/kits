package invoke

import (
	"errors"
	"sync"
)

// ErrDiscoveryNotConfig 出现在没有设置服务发现函数
var ErrDiscoveryNotConfig = errors.New("discovery not config")

// Engine 提供了向服务发送请求的入口
type engine struct {
	dv         DiscoveryFunc
	serviceMap map[string]Service
	mutex      sync.RWMutex
	lbMode     string
	useTracing bool
	useCircuit bool
}

// Init 初始化引擎
func (engine *engine) Init(option *Option) error {
	engine.dv = option.Discover
	engine.lbMode = option.LoadBalanceMode

	return nil
}

// Service 获取一个服务
func (engine *engine) Service(name string) Service {
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
func (engine *engine) Addr(addr string) Service {
	return engine.newAddr(addr)
}

// newAddr 创建一个服务
func (engine *engine) newService(serviceName string, discovery DiscoveryFunc) Service {
	return &service{
		discovery:   discovery,
		name:       serviceName,
		useTracing: engine.useTracing,
		useCircuit: engine.useCircuit,
	}
}

// newAddr 创建固定IP的匿名服务
func (engine *engine) newAddr(addr string) Service {
	discovery := func(string) ([]string, []string, error) {
		return []string{addr}, []string{addr}, nil
	}
	return &service{
		discovery: discovery,
		name:     addr,
	}
}

func newEngine() *engine {
	return &engine{
		dv: func(name string) ([]string, []string, error) {
			return nil, nil, ErrDiscoveryNotConfig
		},
		serviceMap: make(map[string]Service, 10),
	}
}
