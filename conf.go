package gowk

import (
	"os"
	"strconv"
	"strings"
)

var (
	baseURL       = getEnv("BASE_URL", "http://localhost:3030")
	databaseDsn   = getEnv("DATABASE_DSN", "")
	redisAddr     = getEnv("REDIS_ADDR", "")
	redisPassword = getEnv("REDIS_PASSWORD", "")
	redisDB = mustAtoi(getEnv("REDIS_DB", "0"))
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
