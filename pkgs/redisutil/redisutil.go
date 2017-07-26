package redisutil

import (
	"github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/lvhuat/kits/service/profile"
	"strings"
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
