package urpcinvoke

import (
	"sync/atomic"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/lworkltd/kits/service/discovery"
	"github.com/lworkltd/kits/service/restful/code"
)

type Service struct {
	name string
	f    DiscoveryFunc

	useTracing  bool
	useCircuit  bool
	doLogger    bool
	hystrixInfo hystrix.CommandConfig

	selectIndex Counter
}

type Counter int64

func (c *Counter) AddOne() int64 {
	return atomic.AddInt64((*int64)(c), 1)
}

func (service *Service) Call(reqName ...string) *Client {
	cli := &Client{
		serviceName: service.name,
		since:       time.Now().UTC(),
		useCircuit:  service.useCircuit,
		hystrixInfo: service.hystrixInfo,
		useTracing:  service.useTracing,
		doLogger:    service.doLogger,
		discovery:   service.getAddr,
	}
	if len(reqName) > 0 {
		cli.callName = reqName[0]
	}

	return cli
}

func (service *Service) getAddr() (string, string, code.Error) {
	addrs, ids, err := service.f(service.name)
	if err != nil {
		return "", "", nil
	}

	if len(addrs) == 0 {
		return "", "", code.NewMcodef("DISCOVERY_FAILED", "%s not valid node for rpc", service.name)
	}

	idx := int(service.selectIndex.AddOne()) % len(addrs)

	return addrs[idx], ids[idx], nil
}

func cloneAndAppendService(services map[string]*Service, s *Service) map[string]*Service {
	ss := make(map[string]*Service, len(services)+1)
	for name, service := range services {
		ss[name] = service
	}

	ss[s.name] = s

	return ss
}

func getServiceByName(t string) *Service {
	s, exist := services[t]
	if !exist {
		d := defaultOption.Discover
		if d == nil {
			d = discovery.Discover
		}
		s = &Service{
			name:       t,
			f:          d,
			useCircuit: defaultOption.UseCircuit,
			useTracing: defaultOption.UseTracing,
			doLogger:   defaultOption.DoLogger,
			hystrixInfo: hystrix.CommandConfig{
				Timeout:                int(defaultOption.DefaultTimeout / time.Millisecond),
				MaxConcurrentRequests:  defaultOption.DefaultMaxConcurrentRequests,
				RequestVolumeThreshold: defaultOption.DefaultErrorPercentThreshold,
			},
		}
		services = cloneAndAppendService(services, s)
	}

	return s
}

func getServiceByAddr(addr string) *Service {
	s, exist := services[addr]
	if !exist {
		s = &Service{
			name: addr,
			f: func(string) ([]string, []string, error) {
				return []string{addr}, []string{addr}, nil
			},
			useCircuit: defaultOption.UseCircuit,
			useTracing: defaultOption.UseTracing,
			doLogger:   defaultOption.DoLogger,
			hystrixInfo: hystrix.CommandConfig{
				Timeout:                int(defaultOption.DefaultTimeout / time.Millisecond),
				MaxConcurrentRequests:  defaultOption.DefaultMaxConcurrentRequests,
				RequestVolumeThreshold: defaultOption.DefaultErrorPercentThreshold,
			},
		}
		services = cloneAndAppendService(services, s)
	}

	return s
}
