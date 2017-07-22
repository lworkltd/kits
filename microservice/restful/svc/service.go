package svc

import (
	"fmt"
	"time"

	"lwork.com/kits/utils/co"
)

// ServiceImpl 是用来获取服务器地址，并创建调用的
// 此处使用robin作为负载均衡的策略
type ServiceImpl struct {
	wellCount      co.Int64
	totalCallCount co.Int64
	name           string
	disconvery     DiscoveryFunc
}

// 选择服务节点
func (service *ServiceImpl) remote() (string, error) {
	index := service.totalCallCount.Add(1)

	if service.disconvery == nil {
		return "", fmt.Errorf("service %s not found", service.name)
	}

	remotes, err := service.disconvery(service.name)
	if err != nil {
		return "", fmt.Errorf("discovery service %s failed", service.name)
	}

	l := len(remotes)
	if l == 0 {
		return "", fmt.Errorf("service %s not found", service.name)
	}

	if l == 1 {
		return remotes[0], nil
	}

	// 通过轮询策略来访问
	return remotes[index%int64(l)], nil
}

// Get 使用GET方法请求
func (service *ServiceImpl) Get(path string) IClient {
	remote, err := service.remote()
	return newRest(service, "GET", path, remote, err)
}

// Post 使用POST方法请求
func (service *ServiceImpl) Post(path string) IClient {
	remote, err := service.remote()
	return newRest(service, "POST", path, remote, err)
}

// Put 使用PUT方法请求
func (service *ServiceImpl) Put(path string) IClient {
	remote, err := service.remote()
	return newRest(service, "PUT", path, remote, err)
}

// Delete 使用DELETE方法请求
func (service *ServiceImpl) Delete(path string) IClient {
	remote, err := service.remote()
	return newRest(service, "DELETE", path, remote, err)
}

// Method 使用指定方法请求
func (service *ServiceImpl) Method(method, path string) IClient {
	remote, err := service.remote()
	return newRest(service, method, path, remote, err)
}

// Name 返回服务名称
func (service *ServiceImpl) Name() string {
	return service.name
}

func newRest(service IService, method string, path string, remote string, err error) IClient {
	client := &Client{
		createTime:   time.Now(),
		service:      service,
		pathFormat:   path,
		requestHost:  remote,
		sche:         "http",
		errInProcess: err,
		method:       method,
		logFields: map[string]interface{}{
			"service": service.Name(),
			"method":  method,
			"path":    path,
		},
	}

	if err != nil {
		client.logFields["error"] = err
		client.errInProcess = fmt.Errorf("Service not valiable")
	}

	return client
}
