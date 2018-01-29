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
	// 必须配置
	Name     string // *服务名
	Id       string // *服务ID,全局唯一
	Ip       string // *服务端口
	Port     int    // *端口
	CheckUrl string // *HTTP 地址

	// 选项配置
	CheckInterval                string   // 检测间隔，默认 5s
	CheckTimeout                 string   // 检测超时，默认 3s
	CheckDeregisterCriticalAfter string   // 仅支持consul 0.7+
	Tags                         []string // 服务标签
}

// NewConsulClient 创建一个consul客户端
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

// Register 向consul上报一个服务
func (client *Client) Register(option *RegisterOption) error {
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

// Unregister 删除一个服务节点
func (client *Client) Unregister(option *RegisterOption) error {
	return client.cli.Agent().ServiceDeregister(option.Id)
}

// Discover 从consul发现一个服务
func (client *Client) Discover(name string) ([]string, []string, error) {
	client.mutex.RLock()
	service, exist := client.serviceCache[name]
	client.mutex.RUnlock()

	if !exist || service == nil || service.err != nil {
		s, err := client.service(name)
		if err != nil {
			return nil, nil, err
		}

		// 有一定的可能会重复查询
		client.mutex.Lock()
		client.serviceCache[name] = s
		client.mutex.Unlock()
		service = s
	}

	if service.err != nil {
		return nil, nil, fmt.Errorf("Get service %s from consul failed:%v", name, service.err)
	}

	// 记录获取服务信息的时间
	service.r = time.Now()

	return service.hosts, service.ids, nil
}

// KeyValue 从consul获取一个key值
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

func SetClient(client *Client) {
	defaultClient = client
}

func Get() *Client {
	return defaultClient
}
