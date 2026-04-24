package gowk

import (
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	baseURL       = getEnv("BASE_URL", "http://localhost:3030")
	databaseDsn   = getEnv("DATABASE_DSN", "")
	redisAddr     = getEnv("REDIS_ADDR", "")
	redisPassword = getEnv("REDIS_PASSWORD", "")
	redisDB       = mustAtoi(getEnv("REDIS_DB", "0"))
)

// 后台重试参数。均走 time.ParseDuration 解析；未设置、解析失败或 <= 0 时回落到默认值。
// 语义：每轮 attempt 失败后 sleep = min(backoff, max)，随后 backoff *= 2，最终被 max 封顶。
var (
	pgRetryBaseInterval = getEnvDuration("DATABASE_RETRY_BASE_INTERVAL", 2*time.Second)
	pgRetryMaxInterval  = getEnvDuration("DATABASE_RETRY_MAX_INTERVAL", 30*time.Second)
	pgPingTimeout       = getEnvDuration("DATABASE_PING_TIMEOUT", 5*time.Second)

	redisRetryBaseInterval = getEnvDuration("REDIS_RETRY_BASE_INTERVAL", 2*time.Second)
	redisRetryMaxInterval  = getEnvDuration("REDIS_RETRY_MAX_INTERVAL", 30*time.Second)
	redisPingTimeout       = getEnvDuration("REDIS_PING_TIMEOUT", 5*time.Second)
)

var (
	httpServerAddr = getEnv("HTTP_SERVER_ADDR", ":3030")
	grpcServerAddr = getEnv("GRPC_SERVER_ADDR", "")
)

func SetHTTPServerAddr(addr string) { httpServerAddr = addr }
func SetGRPCServerAddr(addr string) { grpcServerAddr = addr }

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return defaultValue
}

func mustAtoi(s string) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return 0
}

func HasRedis() bool { return redisAddr != "" }
func HasGRPC() bool  { return grpcServerAddr != "" }

func BaseURL() string {
	return strings.TrimSuffix(baseURL, "/")
}
