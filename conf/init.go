package conf

import (
	"fmt"
	"os"

	"gopkg.in/ini.v1"
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
	cfgTmp, err := ini.InsensitiveLoad(name)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	cfg = cfgTmp
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
