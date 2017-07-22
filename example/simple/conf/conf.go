package conf

import (
	"github.com/lvhuat/kits/microservice/discovery"
	"github.com/lvhuat/kits/microservice/profile"
)

type Profile struct {
	Base    profile.Base
	Service profile.Service
	Redis   profile.Redis
	Mongo   profile.Mongo
	Mysql   profile.Mysql
	Consul  profile.Consul
}

var config Profile

func Parse(f ...string) error {
	tf := "app.toml"
	if len(f) > 0 {
		tf = f[0]
	}

	plan, meta, err := profile.Parse(tf, &config)
	if err != nil {
		return err
	}

	if plan.ConsulKv {
		discovery.NewConsulClient(config.Consul.Url)
	}
	return nil
}
