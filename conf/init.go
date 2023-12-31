package conf

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

var confpath string

func init() {
	flag.StringVar(&confpath, "c", "conf.ini", "use default conf path")
	flag.Parse()
	Init(confpath)
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
	if cfg.HasSection("db") {
		DB = &DBConf{}
		cfg.Section("db").MapTo(DB)
	} else if cfg.HasSection("mysql") {
		DB = &DBConf{}
		cfg.Section("mysql").MapTo(DB)
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
