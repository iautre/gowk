package gowk

import (
	"os"
	"strconv"
)

var (
	serverAddr       = getEnv("SERVER_ADDR", ":3030")
	databaseDsn      = getEnv("DATABASE_DSN", "")
	redisAddr        = getEnv("REDIS_ADDR", "redis:6379")
	redisPassword    = getEnv("REDIS_PASSWORD", "redispassword")
	redisDB          = mustAtoi(getEnv("REDIS_DB", "0"))
	weappAppid       = getEnv("WEAPP_APPID", "")
	weappSecret      = getEnv("WEAPP_SECRET", "")
	weappJsapiTicket = getEnv("WEAPP_JSAPI_TICKET", "0")
)

// Helper funcs
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
