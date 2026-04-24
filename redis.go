package gowk

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/redis/go-redis/v9"
)

var (
	defaultRedis     atomic.Pointer[redis.Client]
	redisInitOnce    sync.Once
	redisRetryCancel context.CancelFunc
)

// Redis 返回 go-redis 客户端，通过 sync.Once 保证只初始化一次。
// 返回值语义：
//   - REDIS_ADDR 未配置 / 后台尚未 Ping 通 → nil
//   - 后台 Ping 成功后 → 非 nil 客户端
//
// 调用方应检查 nil 并返回错误，不要对 nil 客户端直接调用方法。
func Redis() *redis.Client {
	redisInitOnce.Do(initRedis)
	return defaultRedis.Load()
}

func initRedis() {
	if !HasRedis() {
		return
	}
	// NewClient 只建一次：go-redis 内部维护连接池与后台心跳，反复 New 会积累资源。
	// 这里成功前不 Store 到 defaultRedis，避免外部在未 Ping 通时就拿到一个不可用 client。
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	ctx, cancel := context.WithCancel(context.Background())
	redisRetryCancel = cancel
	go func() {
		retryBackground(ctx, "Redis", redisRetryBaseInterval, redisRetryMaxInterval, func(c context.Context) error {
			pingCtx, cancelPing := context.WithTimeout(c, redisPingTimeout)
			defer cancelPing()
			if err := client.Ping(pingCtx).Err(); err != nil {
				return err
			}
			defaultRedis.Store(client)
			slog.Info("Redis 就绪", "addr", redisAddr)
			return nil
		})
		// ctx 取消路径下如果始终没 Ping 通，defaultRedis 仍为 nil，
		// 这里收尾关闭掉未发布出去的 client，避免 goroutine/连接泄漏。
		if defaultRedis.Load() == nil {
			_ = client.Close()
		}
	}()
}

func closeRedis() {
	if redisRetryCancel != nil {
		redisRetryCancel()
	}
	if c := defaultRedis.Load(); c != nil {
		_ = c.Close()
	}
}
