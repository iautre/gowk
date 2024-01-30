package conf

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	for k, v := range tempConfMap {
		if strings.ToLower(k) == "database" {
			if databases, ok := v.([]map[string]any); ok {
				for _, database := range databases {
					key, d := toDatabaseConf(database)
					dbMap[key] = d
					dbs = append(dbs, dbMap[key])
				}
			} else if database, ok := v.(map[string]any); ok {
				key, d := toDatabaseConf(database)
				dbMap[key] = d
				dbs = append(dbs, dbMap[key])
			}
		} else if strings.ToLower(k) == "server" {
			if temp, ok := v.(map[string]any); ok {
				server = &ServerConf{
					Addr: fmt.Sprintf("%v", temp["addr"]),
				}
			}
		} else if strings.ToLower(k) == "redis" {
			if temp, ok := v.(map[string]any); ok {
				redisConf = toRedisConf(temp)
			}
		} else {
			confMap[k] = v
		}
	}
	dbs = make([]*DatabaseConf, 0, len(dbMap))
	for _, v := range dbMap {
		dbs = append(dbs, v)
	}
}

func toDatabaseConf(database map[string]any) (string, *DatabaseConf) {
	key := "default"
	if tempKey, ok := database["key"]; ok {
		key = tempKey.(string)
	}
	buf, _ := json.Marshal(database)
	var data DatabaseConf
	json.Unmarshal(buf, &data)
	// dbMap[key] = &DatabaseConf{
	// 	Key:         key,
	// 	Type:        database["type"].(string),
	// 	User:        database["user"].(string),
	// 	Password:    database["password"].(string),
	// 	Host:        database["host"].(string),
	// 	Name:        database["name"].(string),
	// 	TablePrefix: database["tablePrefix"].(string),
	// 	Port:        int(database["port"].(int64)),
	// 	MaxPoolSize: uint64(database["maxPoolSize"].(int64)),
	// }
	return key, &data
}

func toRedisConf(redis map[string]any) *ReidsConf {
	buf, _ := json.Marshal(redis)
	var data ReidsConf
	json.Unmarshal(buf, &data)
	return &data
}
