package consul

import (
	"strings"

	"fmt"

	"time"

	"sync"

	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
)

const (
	defaultRegisterCheckInterval = "5s"
	defaultRegisterCheckTimeout  = "3s"
)

var ErrConsulNotInit = errors.New("consul not init yet,please initilize with consul.InitConsul() or consul.SetClient()")

// ConsulClient 服务发现
type Client struct {
	cli          *api.Client
	mutex        sync.RWMutex
	serviceCache map[string]*serviceCache
}

// serviceCache 缓存服务的发现信息
type serviceCache struct {
	t     time.Time
	hosts []string
	ids   []string
	r     time.Time
	err   error
	from  string
}

// RegisterOption 注册服务的选项参数
type RegisterOption struct {
	ServerType ServerType
	// 必须配置
	Name string // *服务名
	Id   string // *服务ID,全局唯一
	Ip   string // *服务端口
	Port int    // *端口
	// 如果是HTTP服务，则为自定义的健康检测的地址，比如`http://10.0.17.90:8080/health`
	// 如果是GRPC服务，则为GRPC服务对象的地址，比如`10.0.17.90:8080/{service}`
	// 	- 如果service为空,则检测服务器所有服务的状态（实际总是返回健康）
	//	- 如果service不为空,则检测服务器指定服务的状态，例如`grpc.health.v1.Health`
	// GRPC 仅支持v1.0.6以上consul版本
	CheckUrl string

	// 选项配置
	CheckInterval                string   // 检测间隔，默认 5s
	CheckTimeout                 string   // 检测超时，默认 3s
	CheckDeregisterCriticalAfter string   // 仅支持consul 0.7+
	Tags                         []string // 服务标签
}

// ServerType 服务类型
type ServerType int32

const (
	// ServerTypeHttp HTTP服务
	ServerTypeHttp = iota
	// ServerTypeGrpc GRPC服务
	ServerTypeGrpc
)

func (serverType ServerType) String() string {
	switch serverType {
	case ServerTypeHttp:
		return "HttpServer"
	case ServerTypeGrpc:
		return "GrpcServer"
	}

	return fmt.Sprintf("UnkownServerType[%d]", serverType)
}

// New 创建一个consul客户端
func New(host string) (*Client, error) {
	if !strings.HasPrefix(host, "http") && !strings.HasPrefix(host, "unix") {
		host = "http://" + host
	}
	cli, err := api.NewClient(&api.Config{
		Address: host,
	})
	if err != nil {
		return nil, err
	}

	consul := &Client{
		cli:          cli,
		serviceCache: make(map[string]*serviceCache, 10),
	}

	go consul.loop()

	return consul, nil
}

func (client *Client) registerHttp(option *RegisterOption) error {
	var tcpStr string
	if "" == option.CheckUrl {
		tcpStr = fmt.Sprintf("%v:%v", option.Ip, option.Port)
	}

	return client.cli.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      option.Id,   // SERVICE_ID
		Name:    option.Name, // 模块定义 fw_service
		Port:    option.Port, // 端口
		Tags:    option.Tags, // 服务标签
		Address: option.Ip,   // 服务地址
		Check: &api.AgentServiceCheck{
			HTTP:     option.CheckUrl,
			TCP:      tcpStr,
			Interval: option.CheckInterval,
			Timeout:  option.CheckTimeout,
			DeregisterCriticalServiceAfter: option.CheckDeregisterCriticalAfter,
		}, // 健康检测
	})
}

func (client *Client) registerGrpc(option *RegisterOption) error {
	return client.cli.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      option.Id,
		Name:    option.Name,
		Port:    option.Port,
		Tags:    option.Tags,
		Address: option.Ip,
		Check: &api.AgentServiceCheck{
			GRPC:     option.CheckUrl,
			Interval: option.CheckInterval,
			Timeout:  option.CheckTimeout,
			DeregisterCriticalServiceAfter: option.CheckDeregisterCriticalAfter,
		}, // 健康检测
	})
}
func checkAndDefaultOption(option *RegisterOption) error {
	if option.Ip == "" {
		return fmt.Errorf("ip must be set in option")
	}

	if option.Port == 0 {
		return fmt.Errorf("port must be set in option")
	}

	if option.Name == "" {
		return fmt.Errorf("service name must be set in option")
	}

	if option.Id == "" {
		return fmt.Errorf("service id must be set in option")
	}

	if option.CheckInterval == "" {
		option.CheckInterval = defaultRegisterCheckInterval
	}

	if option.CheckTimeout == "" {
		option.CheckTimeout = defaultRegisterCheckTimeout
	}

	_, err := time.ParseDuration(option.CheckInterval)
	if err != nil {
		return fmt.Errorf("check interval %s is not a golang duration", option.CheckInterval)
	}

	_, err = time.ParseDuration(option.CheckTimeout)
	if err != nil {
		return fmt.Errorf("check timout %s is not a golang duration", option.CheckTimeout)
	}

	return nil
}

// Register 向consul上报一个服务
func (client *Client) Register(option *RegisterOption) error {
	if err := checkAndDefaultOption(option); err != nil {
		return err
	}

	switch option.ServerType {
	case ServerTypeHttp:
		return client.registerHttp(option)
	case ServerTypeGrpc:
		return client.registerGrpc(option)
	}

	return fmt.Errorf("consul do not support %v", option.ServerType)
}

// Unregister 删除一个服务节点
func (client *Client) Unregister(option *RegisterOption) error {
	return client.cli.Agent().ServiceDeregister(option.Id)
}

// Discover 从consul发现一个服务
func (client *Client) Discover(name string) ([]string, []string, error) {
	var (
		service *serviceCache
		exist   bool
	)
	func() {
		client.mutex.RLock()
		defer client.mutex.RUnlock()
		service, exist = client.serviceCache[name]
	}()

	if !exist || service == nil || service.err != nil {
		s, err := client.service(name)
		if err != nil {
			return nil, nil, err
		}
		// 有一定的可能会重复查询
		func() {
			client.mutex.Lock()
			defer client.mutex.Unlock()

			client.serviceCache[name] = s
		}()
		service = s
	}

	if service.err != nil {
		return nil, nil, fmt.Errorf("Get service %s from consul failed:%v", name, service.err)
	}

	// 记录获取服务信息的时间
	service.r = time.Now()

	return service.hosts, service.ids, nil
}

// KeyValue 从consul获取一个键值
func (client *Client) KeyValue(key string) (string, bool, error) {
	if client == nil || client.cli == nil {
		return "", false, ErrConsulNotInit
	}
	pair, _, err := client.cli.KV().Get(key, nil)

	if err != nil {
		return "", false, err
	}

	if pair == nil {
		return "", false, nil
	}

	return string(pair.Value), true, err
}

// UpdateKeyValue 更见键值
func (client *Client) UpdateKeyValue(key, value string) error {
	if client == nil || client.cli == nil {
		return ErrConsulNotInit
	}
	_, err := client.cli.KV().Put(&api.KVPair{
		Key:   key,
		Value: []byte(value),
	}, nil)

	return err
}

// 循环地去读取已经访问过的服务
func (client *Client) loop() {
	for {
		<-time.After(time.Second * 5)

		// 计算需要更新的服务
		queryServices := make([]string, 0, len(client.serviceCache))
		deleteServices := make([]string, 0, len(client.serviceCache))
		func() {
			client.mutex.RLock()
			defer client.mutex.RUnlock()

			for name, service := range client.serviceCache {
				// 超过3分钟没有访问的服务，将移除自动更新
				if service.r.Add(time.Minute * 3).Before(time.Now()) {
					deleteServices = append(deleteServices, name)
				}

				// 超过5秒没有更新的服务将更新地址
				query := service.t.Add(time.Second * 5).Before(time.Now())
				if query {
					queryServices = append(queryServices, name)
				}
			}
		}()

		client.removeServices(deleteServices)

		if len(queryServices) <= 0 {
			continue
		}

		// 获取服务
		updates := make(map[string]*serviceCache, len(queryServices))
		for _, s := range queryServices {
			newCache, _ := client.service(s)
			updates[s] = newCache
		}

		client.mergeServices(updates)
	}
}

// 获取一个健康的服务
func (client *Client) service(service string) (*serviceCache, error) {
	entrys, _, err := client.cli.Health().Service(service, "", true, nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"service": service,
		}).Warn("Get service from consul failed")

		return &serviceCache{
			err: err,
		}, err
	}

	if len(entrys) == 0 {
		err := fmt.Errorf("not found health service on consul")
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"service": service,
		}).Warn("Get service from consul failed")
		return &serviceCache{
			err: err,
		}, err
	}

	hosts := make([]string, len(entrys))
	ids := make([]string, len(entrys))
	for index, entry := range entrys {
		hosts[index] = fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port)
		ids[index] = entry.Service.ID
	}

	return &serviceCache{
		t:     time.Now(),
		hosts: hosts,
		ids:   ids,
		from:  "consul",
	}, nil
}

// 合并服务信息
func (client *Client) mergeServices(newServices map[string]*serviceCache) {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	for name, info := range newServices {
		client.serviceCache[name] = info
		logrus.WithFields(logrus.Fields{
			"name":  name,
			"hosts": info.hosts,
			"ids":   info.ids,
			"err":   info.err,
		}).Debug("Update service discovery")
	}
}

// 删除服务信息
func (client *Client) removeServices(names []string) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	for _, name := range names {
		logrus.WithFields(logrus.Fields{
			"name": name,
		}).Debug("Remove unuse service")

		delete(client.serviceCache, name)
	}
}

var defaultClient *Client
var doInitConsulOnce sync.Once

// InitConsul 初始化consul
func InitConsul(host string) error {
	cli, err := New(host)
	if err != nil {
		return err
	}

	defaultClient = cli

	return nil
}

// SetClient 从外部传入Consul客户端
func SetClient(client *Client) {
	defaultClient = client
}

func Get() *Client {
	return defaultClient
}
