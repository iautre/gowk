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
	cfgTmp, err := ini.InsensitiveLoad(name)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	cfg = cfgTmp
	Server = &ServerConf{}
	if cfg.HasSection("server") {
		cfg.Section("server").MapTo(Server)
	} else {
		fmt.Printf("no conf server, use default Addr: 80")
		Server.Addr = ":80"
	}
	if cfg.HasSection("mysql") {
		Mysql = &MysqlConf{}
		cfg.Section("mysql").MapTo(Mysql)
	}

	if cfg.HasSection("mongo") {
		Mongo = &MongoConf{}
		cfg.Section("mongo").MapTo(Mongo)
	}

	//组合全部
	confs = make(conf)
	for _, se := range cfg.Sections() {
		for _, key := range se.Keys() {
			confs[fmt.Sprintf("%s.%s", se.Name(), key.Name())] = key.Value()
		}
	}
}
