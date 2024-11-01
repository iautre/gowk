package conf

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

var confpath string

func init() {
	flag.StringVar(&confpath, "c", "conf.toml", "use default conf path")
	testing.Init()
	flag.Parse()
	Init(confpath)
}

func Init(path string) {
	var tempConfMap map[string]any
	if _, err := toml.DecodeFile(path, &tempConfMap); err != nil {
		fmt.Printf("Fail to read conf file: %v \n", err)
		// os.Exit(1)
	}
	for k, v := range tempConfMap {
		if strings.ToLower(k) == "database" {
			if temp, ok := v.(map[string]any); ok {
				db = map2Struct[DatabaseConf](temp)
			}
		} else if strings.ToLower(k) == "server" {
			if temp, ok := v.(map[string]any); ok {
				server = map2Struct[ServerConf](temp)
			}
		} else if strings.ToLower(k) == "redis" {
			if temp, ok := v.(map[string]any); ok {
				redisConf = map2Struct[RedisConf](temp)
			}
		} else if strings.ToLower(k) == "weapp" {
			if temp, ok := v.(map[string]any); ok {
				weappConf = map2Struct[WeappConf](temp)
			}
		} else {
			confMap[k] = v
		}
	}
	if server == nil {
		server = &ServerConf{
			Addr: ":8080",
		}
	}
	setEnv2Conf(server)
}

func map2Struct[T any](data map[string]any) *T {
	buf, _ := json.Marshal(data)
	var res T
	json.Unmarshal(buf, &res)
	return &res
}

func setEnv2Conf[T ServerConf | RedisConf | DatabaseConf](t *T) {
	ty := reflect.TypeOf(t).Elem()
	tv := reflect.ValueOf(t).Elem()
	for i := 0; i < ty.NumField(); i++ {
		field := ty.Field(i)
		// 获取结构体类型名和字段名
		typeName := strings.ReplaceAll(strings.ToUpper(ty.Name()), "CONF", "")
		fieldName := strings.ToUpper(field.Name)

		// 组合成字符串，例如：ServerConf_Addr
		key := typeName + "_" + fieldName
		if val := os.Getenv(key); val != "" {
			tv.Field(i).SetString(val)
		}
	}
}
