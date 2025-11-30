package gowk

import (
	"github.com/redis/go-redis/v9"
	"sync"
)

var defaultRedis *redis.Client
var redisInitOnce sync.Once

func Redis() *redis.Client {
	redisInitOnce.Do(initRedis)
	return defaultRedis
}

func initRedis() {
	if HasRedis() {
		defaultRedis = redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: redisPassword, // 没有密码，默认值
			DB:       redisDB,       // 默认DB 0
		})
	}
}
