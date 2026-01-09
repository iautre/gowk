package gowk

import (
	"os"
	"strconv"
	"strings"
)

var (
	baseURL          = getEnv("BASE_URL", "http://192.168.199.60:8087")
	httpServerAddr   = getEnv("HTTP_SERVER_ADDR", ":8087")
	grpcServerAddr   = getEnv("GRPC_SERVER_ADDR", ":50051")
	databaseDsn      = getEnv("DATABASE_DSN", "")
	redisAddr        = getEnv("REDIS_ADDR", "")
	redisPassword    = getEnv("REDIS_PASSWORD", "")
	redisDB          = mustAtoi(getEnv("REDIS_DB", "0"))
	weappAppid       = getEnv("WEAPP_APPID", "")
	weappSecret      = getEnv("WEAPP_SECRET", "")
	weappJsapiTicket = getEnv("WEAPP_JSAPI_TICKET", "0")
	authAPIPrefix    = getEnv("AUTH_API_PREFIX", "")
)

// Helper funcs
func getEnv(key, defaultValue string) string {
	// 从环境变量中获取指定键的值
	if v := os.Getenv(key); v != "" {
		// 如果环境变量中存在该键且值不为空，则返回该值
		return v
	}
	// 如果环境变量中不存在该键或值为空，则返回默认值
	return defaultValue
}

func mustAtoi(s string) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return 0
}

func HasRedis() bool {
	return redisAddr != ""
}

func HasWeapp() bool {
	return weappAppid != "" && weappSecret != ""
}

func AuthAPIPrefix() string {
	prefix := authAPIPrefix
	if prefix == "" {
		return ""
	}
	// Remove trailing slash
	prefix = strings.TrimSuffix(prefix, "/")
	// Add leading slash if missing
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	return prefix
}

func BaseURL() string {
	url := baseURL
	if url == "" {
		return ""
	}
	// Remove trailing slash for consistency
	return strings.TrimSuffix(url, "/")
}
