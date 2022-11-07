package redis

import (
	"strings"

	"github.com/go-redis/redis"
	"github.com/lworkltd/kits/service/profile"
	"github.com/sirupsen/logrus"
)

func validAddr(addr string) bool {
	return addr != ""
}

func Option(cfg *profile.Redis) *redis.ClusterOptions {
	cfgAddrs := strings.Split(cfg.Endpoints, ",")
	addrs := make([]string, 0, len(cfgAddrs))
	for _, addr := range cfgAddrs {
		if !validAddr(addr) {
			continue
		}
		addrs = append(addrs, addr)
	}

	logrus.WithField("redis-addrs", addrs).Debug("New redis option")

	return &redis.ClusterOptions{
		Addrs:    addrs,
		Password: cfg.Password,
	}
}
