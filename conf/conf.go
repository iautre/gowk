package conf

import (
	"encoding/json"
	"time"
)

type DatabaseConf struct {
	Type            string        `json:"type" toml:"type"`
	User            string        `json:"user" toml:"user"`
	Password        string        `json:"password" toml:"password"`
	Host            string        `json:"host" toml:"host"`
	Name            string        `json:"name" toml:"name"`
	TablePrefix     string        `json:"table_prefix" toml:"table_prefix"`
	Port            int           `json:"port" toml:"port"`
	MaxPoolSize     uint64        `json:"max_pool_size" toml:"max_pool_size"`
	MaxOpenConns    int           `json:"max_open_conns" toml:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns" toml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" toml:"conn_max_lifetime"`
}
type ServerConf struct {
	Addr   string `toml:"addr"`
	Cert   string `toml:"cert"`
	Key    string `toml:"key"`
	Domain string `toml:"domain"`
}
type RedisConf struct {
	Host     string `json:"host" toml:"host"`
	Port     int    `json:"port" toml:"port"`
	Password string `json:"password" toml:"password"`
	DB       int    `json:"db" toml:"db"`
}
type WeappConf struct {
	Appid       string `json:"appid" toml:"appid"`
	Secret      string `json:"secret" toml:"secret"`
	JsapiTicket bool   `json:"jsapi_ticket" toml:"jsapi_ticket"`
}

var (
	confMap   map[string]any = map[string]any{}
	db        *DatabaseConf
	server    *ServerConf
	redisConf *RedisConf
	weappConf *WeappConf
)

func Redis() *RedisConf {
	return redisConf
}
func HasRedis() bool {
	return redisConf != nil
}
func Weapp() *WeappConf { return weappConf }
func HasWeapp() bool {
	return weappConf != nil
}
func Server() *ServerConf {
	return server
}
func DB() *DatabaseConf {
	return db
}

func HasDB() bool {
	return db != nil
}
func Get[T any](key string) *T {
	if _, ok := confMap[key]; !ok {
		return nil
	}
	if m, ok := confMap[key].(map[string]any); ok {
		buf, _ := json.Marshal(m)
		var data T
		json.Unmarshal(buf, &data)
		return &data
	} else if ms, ok := confMap[key].([]map[string]any); ok {
		if len(ms) > 0 {
			buf, _ := json.Marshal(ms[0])
			var data T
			json.Unmarshal(buf, &data)
			return &data
		}
	}
	return nil
}
func Gets[T any](key string) []*T {
	if _, ok := confMap[key]; !ok {
		return nil
	}
	if m, ok := confMap[key].(map[string]any); ok {
		buf, _ := json.Marshal(m)
		var data T
		json.Unmarshal(buf, &data)
		res := make([]*T, 0, 1)
		return append(res, &data)
	} else if ms, ok := confMap[key].([]map[string]any); ok {
		res := make([]*T, 0, len(ms))
		for _, m := range ms {
			buf, _ := json.Marshal(m)
			var data T
			json.Unmarshal(buf, &data)
			res = append(res, &data)
		}
		return res
	}
	return nil
}
