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

	// pgEnabled 来自 DATABASE_ENABLED，默认 true（严格模式）：必须配置 DATABASE_DSN
	// 且在启动期内连接成功，否则进程不启动。只有显式设为 false 才允许在没有
	// PostgreSQL 的情况下运行，此时 Postgres(ctx) 恒为 nil。
	pgEnabled = getEnvBool("DATABASE_ENABLED", true)

	// cookieSecure 决定浏览器会话 cookie（oidc_jwt）是否带 Secure 标记。
	// COOKIE_SECURE 优先；未设置时按 BASE_URL 的协议推导——http 部署下必须为
	// false，否则浏览器会直接丢弃 cookie，登录态永远写不进去。
	cookieSecure = getEnvBool("COOKIE_SECURE", !strings.HasPrefix(BaseURL(), "http://"))
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

	// gRPC 健康状态刷新间隔：后台按此周期把"已配置依赖是否就绪"同步到 grpc.health.v1.Health。
	grpcHealthRefreshInterval = getEnvDuration("GRPC_HEALTH_REFRESH_INTERVAL", 2*time.Second)

	// 数据库迁移建立独立连接的总超时：在此时间内带重试连接 DB，超时仍连不上则迁移失败（fail-fast）。
	migrationConnectTimeout = getEnvDuration("DATABASE_MIGRATION_CONNECT_TIMEOUT", 30*time.Second)
)

var (
	httpServerAddr = getEnv("HTTP_SERVER_ADDR", ":3030")
	grpcServerAddr = getEnv("GRPC_SERVER_ADDR", "")
)

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

// getEnvBool 读取布尔开关，语义与 getEnv 一致：未设置或无法解析时用默认值。
// 接受 strconv.ParseBool 的全部写法（1/0、t/f、true/false 及其大小写变体）。
func getEnvBool(key string, defaultValue bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
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

// CookieSecure 返回会话 cookie 是否带 Secure 标记。
func CookieSecure() bool { return cookieSecure }

// SetCookieSecure 覆盖 cookie 的 Secure 标记，供测试与特殊部署形态使用。
func SetCookieSecure(secure bool) { cookieSecure = secure }
