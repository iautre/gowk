package gowk

import (
	"github.com/redis/go-redis/v9"
)

var defaultRedis *redis.Client

func Redis() *redis.Client {
	if defaultRedis == nil {
		initRedis()
	}
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
