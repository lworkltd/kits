package invoke

import (
	"fmt"
	"time"

	"github.com/lvhuat/kits/pkgs/co"
)

// service 是用来获取服务器地址，并创建调用的
// 此处使用robin作为负载均衡的策略
type service struct {
	wellCount      co.Int64
	totalCallCount co.Int64
	name           string
	discover       DiscoveryFunc
	useTracing     bool
	useCircuit     bool
}

// 选择服务节点
func (service *service) remote() (string, string, error) {
	index := service.totalCallCount.Get()
	defer service.totalCallCount.Add(1)

	if service.discover == nil {
		return "", "", fmt.Errorf("service %s not found", service.name)
	}

	remotes, ids, err := service.discover(service.name)
	if err != nil {
		return "", "", fmt.Errorf("discovery service %s failed", service.name)
	}

	l := len(remotes)
	if l == 0 {
		return "", "", fmt.Errorf("service %s not found", service.name)
	}

	if l == 1 {
		return remotes[0], ids[0], nil
	}

	use := index % int64(l)
	// 通过轮询策略来访问
	return remotes[use], ids[use], nil
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
	remote, id, err := service.remote()
	return newRest(service, method, path, remote, id, err)
}

// Name 返回服务名称
func (service *service) Name() string {
	return service.name
}

func newRest(service Service, method string, path string, remote string, id string, err error) Client {
	client := &client{
		createTime:   time.Now(),
		service:      service,
		path:         path,
		host:         remote,
		serverid:     id,
		sche:         "http",
		errInProcess: err,
		method:       method,
		logFields: map[string]interface{}{
			"service":   service.Name(),
			"serviceid": id,
			"method":    method,
			"path":      path,
		},
	}

	if err != nil {
		client.logFields["error"] = err
		client.errInProcess = fmt.Errorf("Service not valiable")
	}

	return client
}
