package conf

import (
	"fmt"
	"os"

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
type MysqlConf struct {
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
	Mysql  *MysqlConf
	Mongo  *MongoConf
	Server *ServerConf
)

func init() {
	Init("conf.ini")
}

func Init(name string) {
	// 文件路径
	// if !filepath.IsAbs(name) {
	// 	e, _ := os.Executable()
	// 	name = filepath.Join(filepath.Dir(e), name)
	// }
	cfg, err := ini.InsensitiveLoad(name)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	if cfg.HasSection("mysql") {
		Mysql = &MysqlConf{}
		cfg.Section("mysql").MapTo(Mysql)
	}

	if cfg.HasSection("mongo") {
		Mongo = &MongoConf{}
		cfg.Section("mongo").MapTo(Mongo)
	}
	if cfg.HasSection("server") {
		Server = &ServerConf{}
		cfg.Section("server").MapTo(Server)
	}
	//组合全部
	confs = make(conf)
	for _, se := range cfg.Sections() {
		for _, key := range se.Keys() {
			confs[fmt.Sprintf("%s.%s", se.Name(), key.Name())] = key.Value()
		}
	}
}
