package conf

import (
	"fmt"

	"gopkg.in/ini.v1"
)

type conf map[string]any

type MongoConf struct {
	User        string `ini:"user"`
	Password    string `ini:"password"`
	Host        string `ini:"host"`
	Name        string `ini:"name"`
	TablePrefix string `ini:"table_prefix"`
	Port        int    `ini:"port"`
	MaxPoolSize uint64 `ini:"max_pool_size"`
}
type DBConf struct {
	Type        string `ini:"type"`
	User        string `ini:"user"`
	Password    string `ini:"password"`
	Host        string `ini:"host"`
	Name        string `ini:"name"`
	TablePrefix string `ini:"table_prefix"`
	Port        int    `ini:"port"`
	MaxPoolSize uint64 `ini:"max_pool_size"`
}
type ServerConf struct {
	Addr string `ini:"addr"`
}

var (
	confs  conf
	cfg    *ini.File
	DB     *DBConf
	Mongo  *MongoConf
	Server *ServerConf
)

func Section(name string) *ini.Section {
	return cfg.Section(name)
}

func Get(key string) any {
	return confs[key]
}
func GetString(key string) string {
	return fmt.Sprintf("%v", Get(key))
}
