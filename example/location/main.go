package main

import (
	"github.com/go-redis/redis"
	"github.com/lvhuat/kits/example/location/api/server"
	"github.com/lvhuat/kits/example/location/conf"
	"github.com/lvhuat/kits/example/location/model"
	"github.com/lvhuat/kits/pkgs/redisutil"
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
