package model

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/lworkltd/kits/example/location/api/errcode"
	"github.com/lworkltd/kits/service/restful/code"
	"strconv"
	"sync"
)

type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

var onceInitSession sync.Once
var redisSession *RedisSession

type RedisSession struct {
	*redis.ClusterClient
}

func stringToFloat(v string, cadidate float64) float64 {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return cadidate
	}
	return f
}

func (sess *RedisSession) GetCitizenLocation(id string) (Location, code.Error) {
	var location Location
	cmd := sess.HGetAll(fmt.Sprintf("citizen.%s", id))
	v, err := cmd.Result()
	if err != nil {
		return Location{}, code.NewError(errcode.ReadCacheFailed, err)
	}
	location.Latitude = stringToFloat(v["latitude"], 0)
	location.Latitude = stringToFloat(v["latitude"], 0)

	return location, nil
}

func Setup(clusterClient *redis.ClusterClient) {
	onceInitSession.Do(func() {
		redisSession = &RedisSession{
			ClusterClient: clusterClient,
		}
	})
}

func GetRedisSession() *RedisSession {
	return redisSession
}
