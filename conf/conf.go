package conf

import (
	"fmt"
)

type DatabaseConf struct {
	Key         string `json:"key" toml:"key"`
	Type        string `json:"type" toml:"type"`
	User        string `json:"user" toml:"user"`
	Password    string `json:"password" toml:"password"`
	Host        string `json:"host" toml:"host"`
	Name        string `json:"name" toml:"name"`
	TablePrefix string `json:"table_prefix" toml:"table_prefix"`
	Port        int    `json:"port" toml:"port"`
	MaxPoolSize uint64 `json:"max_pool_size" toml:"max_pool_size"`
}
type ServerConf struct {
	Addr string `toml:"addr"`
}

var (
	confMap map[string]any           = map[string]any{}
	dbMap   map[string]*DatabaseConf = map[string]*DatabaseConf{}
	dbs     []*DatabaseConf
	server  *ServerConf
)

func Server() *ServerConf {
	return server
}
func DB(key string) *DatabaseConf {
	return dbMap[key]
}
func DBs() []*DatabaseConf {
	return dbs
}
func HasDB() bool {
	return len(dbs) > 0
}
func Get(key string) any {
	return confMap[key]
}
func GetString(key string) string {
	return fmt.Sprintf("%v", Get(key))
}
