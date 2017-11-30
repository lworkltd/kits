package invoke

import (
	"fmt"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/lworkltd/kits/utils/co"
)

// service 是用来获取服务器地址，并创建调用的
// 此处使用robin作为负载均衡的策略
type service struct {
	wellCount      co.Int64
	totalCallCount co.Int64
	name           string
	discovery      DiscoveryFunc
	useTracing     bool
	useCircuit     bool
	circuitConfig  hystrix.CommandConfig
}

// 选择服务节点
func (service *service) remote() (string, string, error) {
	index := service.totalCallCount.Get()
	defer service.totalCallCount.Add(1)

	if service.discovery == nil {
		return "", "", fmt.Errorf("service %s not found", service.name)
	}

	remotes, ids, err := service.discovery(service.name)
	if err != nil {
		return "", "", fmt.Errorf("discovery service %s failed", service.name)
	}

	if len(remotes) != len(ids) {
		return "", "", fmt.Errorf("discovery return wrong remotes=%v ids=%v", remotes, ids)
	}

	l := len(remotes)
	if l == 0 {
		return "", "", fmt.Errorf("service %s not found", service.name)
	}
	var use int64
	if l > 1 {
		// 通过轮询策略来访问
		// TODO:支持更多的策略
		use = index % int64(l)
	}
	addr, id := remotes[use], ids[use]

	if service.useCircuit {
		if _, exist, _ := hystrix.GetCircuit(id); !exist {
			hystrix.ConfigureCommand(id, service.circuitConfig)
		}
	}

	return addr, id, nil
}

// Get 使用GET方法请求
func (service *service) Get(path string) Client {
	return service.Method("GET", path)
}

// Post 使用POST方法请求
func (service *service) Post(path string) Client {
	return service.Method("POST", path)
}

// Put 使用PUT方法请求
func (service *service) Put(path string) Client {
	return service.Method("PUT", path)
}

// Delete 使用DELETE方法请求
func (service *service) Delete(path string) Client {
	return service.Method("DELETE", path)
}

// Method 使用指定方法请求
func (service *service) Method(method, path string) Client {
	return newRest(service, method, path)
}

func (service *service) Remote() (string, string, error) {
	return service.remote()
}

// Name 返回服务名称
func (service *service) Name() string {
	return service.name
}

func newRest(service Service, method string, path string) Client {
	client := &client{
		createTime: time.Now(),
		service:    service,
		path:       path,
		scheme:     "http",
		method:     method,
		logFields: map[string]interface{}{
			"service": service.Name(),
			"method":  method,
			"path":    path,
		},
	}

	return client
}
