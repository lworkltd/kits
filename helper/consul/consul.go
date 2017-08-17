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

var ErrConsulNotInit = errors.New("consul not init yet,please initilize with consul.InitConsul() or consul.SetClient()")

// ConsulClient 服务发现
type Client struct {
	cli          *api.Client
	mutex        sync.RWMutex
	serviceCache map[string]*serviceCache
}

type serviceCache struct {
	t     time.Time
	hosts []string
	ids   []string
	r     time.Time
	err   error
	from  string
}

type RegisterOption struct {
	Ip       string
	Port     int
	CheckUrl string
	Name     string
	Id       string
	Tags     []string
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
	return client.cli.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      option.Id,   // SERVICE_ID
		Name:    option.Name, // 模块定义 fw_service
		Port:    option.Port, // 端口
		Tags:    option.Tags, // 服务标签
		Address: option.Ip,   // 服务地址
		Check: &api.AgentServiceCheck{
			HTTP: option.CheckUrl,
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

	if !exist {
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

// 循环地去读取已经访问过的服务
func (client *Client) loop() {
	for {
		<-time.After(time.Minute)

		// 计算需要更新的服务
		queryServices := make([]string, 0, len(client.serviceCache))
		deleteServices := make([]string, 0, len(client.serviceCache))
		for name, service := range client.serviceCache {
			// 超过30分钟没有访问的服务，将移除自动更新
			if service.r.Add(time.Minute * 30).Before(time.Now()) {
				deleteServices = append(deleteServices, name)
			}

			// 超过1分钟没有更新的服务将更新地址
			query := service.t.Add(time.Minute).Before(time.Now())
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
		service := client.serviceCache[name]
		if service == nil {
			service = &serviceCache{}
		}
		service.hosts = info.hosts
		service.t = info.t
		service.err = info.err
		client.serviceCache[name] = service
	}
}

// 删除服务信息
func (client *Client) removeServices(names []string) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	for _, name := range names {
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
