package conf

import (
	"fmt"
)

type conf struct {
	db     []*DatabaseConf `toml:"database"`
	server *ServerConf     `toml:"server"`
	other  map[string]any  `toml:"other"`
}
type DatabaseConf struct {
	Key         string `toml:"key"`
	Type        string `toml:"type"`
	User        string `toml:"user"`
	Password    string `toml:"password"`
	Host        string `toml:"host"`
	Name        string `toml:"name"`
	TablePrefix string `toml:"table_prefix"`
	Port        int    `toml:"port"`
	MaxPoolSize uint64 `toml:"max_pool_size"`
}
type ServerConf struct {
	Addr string `toml:"addr"`
}

var (
	confs conf
	dbMap map[string]*DatabaseConf
)

func Server() *ServerConf {
	return confs.server
}
func DB(key string) *DatabaseConf {
	return dbMap[key]
}
func DBs() []*DatabaseConf {
	return confs.db
}
func HasDB() bool {
	return confs.db != nil && len(confs.db) > 0
}
func Get(key string) any {
	return confs.other[key]
}
func GetString(key string) string {
	return fmt.Sprintf("%v", Get(key))
}
