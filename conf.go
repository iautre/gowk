package gowk

import (
	"os"
	"strconv"
)

var (
	SERVER_ADDR    = getEnvWithDefault("SERVER_ADDR", ":3030")
	DATABASE_DSN   = getEnvWithDefault("DATABASE_DSN", "")
	REDIS_ADDR     = getEnvWithDefault("REDIS_ADDR", "redis:6379")
	REDIS_PASSWORD = getEnvWithDefault("REDIS_PASSWORD", "redispassword")
	REDIS_DB, _    = strconv.Atoi(getEnvWithDefault("REDIS_DB", "0"))
)
var (
	WEAPP_APPID        = getEnvWithDefault("WEAPP_APPID", "")
	WEAPP_SECRET       = getEnvWithDefault("WEAPP_SECRET", "")
	WEAPP_JSAPI_TICKET = getEnvWithDefault("WEAPP_JSAPI_TICKET", "0")
)

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
func HasRedis() bool {
	return REDIS_ADDR != ""
}
func HasWeapp() bool {
	return WEAPP_APPID != ""
}
