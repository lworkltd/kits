package discover

type StaticService struct {
	Name  string
	Hosts []string
}

// StaticDiscoverer 静态服务发现，比如写死在文件里面的
type StaticDiscoverer struct {
	serviceCache map[string]*StaticService
}

// Discover 从consul发现一个服务
func (staticDiscoverer *StaticDiscoverer) Discover(service string) ([]string, []string, error) {
	s, exist := staticDiscoverer.serviceCache[service]
	if !exist {
		return []string{}, []string{}, nil
	}

	return s.Hosts, s.Hosts, nil
}

// NewStaticDiscoverer 创建一个服务实例
func NewStaticDiscoverer(services []*StaticService) *StaticDiscoverer {
	serviceCache := make(map[string]*StaticService, len(services))

	for _, s := range services {
		serviceCache[s.Name] = s
	}

	return &StaticDiscoverer{
		serviceCache: serviceCache,
	}
}
