package gowk

import (
	"fmt"

	"github.com/iautre/gowk/conf"
	"github.com/redis/go-redis/v9"
)

var defaultRedis *redis.Client

func Redis() *redis.Client {
	return defaultRedis
}

func initRedis() {
	if conf.HasRedis() {
		defaultRedis = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%v:%d", conf.Redis().Host, conf.Redis().Port),
			Password: conf.Redis().Password, // 没有密码，默认值
			DB:       conf.Redis().DB,       // 默认DB 0
		})
	}
}
