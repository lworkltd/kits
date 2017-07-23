package conf

import (
	"strings"

	"github.com/lvhuat/kits/microservice/discovery"
	"github.com/lvhuat/kits/microservice/profile"
)

type Profile struct {
	Base      profile.Base
	Service   profile.Service
	Redis     profile.Redis
	Mongo     profile.Mongo
	Mysql     profile.Mysql
	Consul    profile.Consul
	Svc       profile.Svc
	Logger    profile.Logger
	Hystrix   profile.Hystrix
	Zipkin    profile.Zipkin
	Discovery profile.Discovery
}

var config Profile

func Parse(f ...string) error {
	tf := "app.toml"
	if len(f) > 0 {
		tf = f[0]
	}

	plan, _, err := profile.Parse(tf, &config)
	if err != nil {
		return err
	}

	var consulClient *discovery.ConsulClient
	if plan.ConsulKv {
		consulClient, err = discovery.NewConsulClient(config.Consul.Url)
		if err != nil {
			return err
		}
		discovery.SetDefault(consulClient)
	}

	staticsService, err := parseStaticService(config.Discovery.StaticServices)
	if err != nil {
		return err
	}

	var discoveryOption discovery.Option
	if config.Discovery.EnableConsul {
		discoveryOption.ConsulClient = consulClient
	}
	if config.Discovery.EnableStatic {
		discoveryOption.StaticServices = staticsService
	}

	if err := discovery.Init(&discoveryOption); err != nil {
		return err
	}

	return nil
}

func parseStaticService(lines []string) ([]*discovery.StaticService, error) {
	var staticServices []*discovery.StaticService
	for _, line := range lines {
		words := strings.Split(line, " ")
		if len(words) == 0 {

		}
		service := &discovery.StaticService{
			Name:  words[0],
			Hosts: words[1:],
		}

		staticServices = append(staticServices, service)
	}

	return staticServices, nil
}
