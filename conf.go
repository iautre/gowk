package gowk

import (
	"os"
	"strconv"
	"strings"
)

var (
	baseURL          = getEnv("BASE_URL", "http://localhost:3030")
	databaseDsn      = getEnv("DATABASE_DSN", "")
	redisAddr        = getEnv("REDIS_ADDR", "")
	redisPassword    = getEnv("REDIS_PASSWORD", "")
	redisDB          = mustAtoi(getEnv("REDIS_DB", "0"))
	weappAppid       = getEnv("WEAPP_APPID", "")
	weappSecret      = getEnv("WEAPP_SECRET", "")
	weappJsapiTicket = getEnvBool("WEAPP_JSAPI_TICKET", false)
	authAPIPrefix    = getEnv("AUTH_API_PREFIX", "")
)

var (
	httpServerAddr = getEnv("HTTP_SERVER_ADDR", ":3030")
	grpcServerAddr = getEnv("GRPC_SERVER_ADDR", ":3031")
)

func SetHTTPServerAddr(addr string) { httpServerAddr = addr }
func SetGRPCServerAddr(addr string) { grpcServerAddr = addr }

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return defaultValue
	}
	return b
}

func mustAtoi(s string) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return 0
}

func HasRedis() bool { return redisAddr != "" }

func HasWeapp() bool { return weappAppid != "" && weappSecret != "" }

func AuthAPIPrefix() string {
	prefix := authAPIPrefix
	if prefix == "" {
		return ""
	}
	prefix = strings.TrimSuffix(prefix, "/")
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	return prefix
}

func BaseURL() string {
	return strings.TrimSuffix(baseURL, "/")
}
