package invokeimpl

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/lworkltd/kits/service/grpcinvoke"
)

// ErrDiscoveryNotConfig 出现在没有设置服务发现函数
var ErrDiscoveryNotConfig = errors.New("discovery not config")

// Engine 提供了向服务发送请求的入口
type engine struct {
	dv         grpcinvoke.DiscoveryFunc
	lbMode     string
	useTracing bool
	useCircuit bool

	hystrixInfo hystrix.CommandConfig
	mutex       sync.RWMutex
	serviceMap  map[string]grpcinvoke.Service
}

// Init 初始化引擎
func (engine *engine) Init(option *grpcinvoke.Option) error {
	engine.dv = option.Discover
	engine.lbMode = option.LoadBalanceMode
	engine.useTracing = option.UseTracing
	engine.useCircuit = option.UseCircuit
	engine.hystrixInfo.ErrorPercentThreshold = option.DefaultErrorPercentThreshold
	engine.hystrixInfo.MaxConcurrentRequests = option.DefaultMaxConcurrentRequests
	engine.hystrixInfo.Timeout = int(option.DefaultTimeout / time.Millisecond)
	return nil
}

func (engine *engine) Service(name string) grpcinvoke.Service {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	if engine.serviceMap == nil {
		engine.serviceMap = make(map[string]grpcinvoke.Service, 1)
	}
	service, exsit := engine.serviceMap[name]

	if !exsit {
		service = engine.newService(name, engine.dv, false)
		engine.serviceMap[name] = service
	}

	return service
}

func (engine *engine) Addr(addr string) grpcinvoke.Service {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	if engine.serviceMap == nil {
		engine.serviceMap = make(map[string]grpcinvoke.Service, 1)
	}
	service, exsit := engine.serviceMap[addr]

	if !exsit {
		service = engine.newService(addr,
			func(string) ([]string, []string, error) {
				return []string{addr}, []string{addr}, nil
			},
			false,
		)
		engine.serviceMap[addr] = service
	}

	return service
}

func (engine *engine) removeGrpcService(serviceName string) error {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	_, exist := engine.serviceMap[serviceName]
	if !exist {
		return fmt.Errorf("not exist")
	}

	delete(engine.serviceMap, serviceName)

	return nil
}

// newGrpcService 创建一个GRPC调用服务代理实例
func (engine *engine) newService(serviceName string, discovery grpcinvoke.DiscoveryFunc, freeConnAfterUsed bool) grpcinvoke.Service {
	return &grpcService{
		name:              serviceName,
		freeConnAfterUsed: freeConnAfterUsed,
		useTracing:        engine.useTracing,
		useCircuit:        engine.useCircuit,
		connLb:            newGrpcConnBalancer(serviceName, 4, discovery),
		hystrixInfo:       engine.hystrixInfo,
		remove:            func() { engine.removeGrpcService(serviceName) },
	}
}

// newGrpcAddr 创建一个GRPC调用服务代理实例
func (engine *engine) newAddr(addr string, freeConnAfterUsed bool) grpcinvoke.Service {
	return &grpcService{
		name:              addr,
		freeConnAfterUsed: freeConnAfterUsed,
		useTracing:        engine.useTracing,
		useCircuit:        engine.useCircuit,
		hystrixInfo:       engine.hystrixInfo,
		connLb:            newGrpcConnBalancer(addr, 4, createAddrDiscovery(addr)),
		remove:            func() { engine.removeGrpcService(addr) },
	}
}

func newEngine() *engine {
	return &engine{
		dv: func(name string) ([]string, []string, error) {
			return nil, nil, ErrDiscoveryNotConfig
		},
	}
}

func init() {
	grpcinvoke.DefaultEngine = newEngine()
}
