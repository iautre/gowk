package conf

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
)

var confpath string

func init() {
	flag.StringVar(&confpath, "c", "conf.ini", "use default conf path")
	if flag.Lookup("test.v") != nil {
		testing.Init()
	}
	flag.Parse()
	Init(confpath)
}

func Init(path string) {
	if _, err := toml.DecodeFile(path, &confs); err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	dbMap = make(map[string]*DatabaseConf)
	if confs.db != nil && len(confs.db) > 0 {
		for _, v := range confs.db {
			dbMap[v.Key] = v
		}
	}
}
