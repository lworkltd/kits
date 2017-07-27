package main

import (
	"github.com/go-redis/redis"
	"github.com/lworkltd/kits/example/location/api/server"
	"github.com/lworkltd/kits/example/location/conf"
	"github.com/lworkltd/kits/example/location/model"
	"github.com/lworkltd/kits/pkgs/redisutil"
)

func main() {
	if err := conf.Parse(); err != nil {
		panic(err)
	}

	conf.Dump()

	client := redis.NewClusterClient(redisutil.Option(conf.GetRedis()))
	model.Setup(client)

	if err := server.Setup(conf.GetService()); err != nil {
		panic(err)
	}
}
