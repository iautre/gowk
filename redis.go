package gowk

import (
	"context"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"
)

var defaultRedis *redis.Client
var redisInitOnce sync.Once

func Redis() *redis.Client {
	redisInitOnce.Do(initRedis)
	return defaultRedis
}

func initRedis() {
	if !HasRedis() {
		return
	}
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		slog.Error("Redis 连接失败", "addr", redisAddr, "err", err)
		return
	}
	defaultRedis = client
	slog.Info("Redis 连接成功", "addr", redisAddr)
}
