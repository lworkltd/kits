package discovery

type StaticService struct {
	Name  string
	Hosts []string
}

// StaticDiscovery 静态服务发现，比如写死在文件里面的
type StaticDiscovery struct {
	serviceCache map[string]*StaticService
}

// Discover 从consul发现一个服务
func (staticDiscovery *StaticDiscovery) Discover(service string) ([]string, []string, error) {
	s, exist := staticDiscovery.serviceCache[service]
	if !exist {
		return []string{}, []string{}, nil
	}

	return s.Hosts, s.Hosts, nil
}

// NewStaticDiscovery 创建一个服务实例
func NewStaticDiscovery(services []*StaticService) *StaticDiscovery {
	serviceCache := make(map[string]*StaticService, len(services))

	for _, service := range services {
		serviceCache[service.Name] = service
	}

	return &StaticDiscovery{
		serviceCache: serviceCache,
	}
}
