package gowk

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

type conf struct {
	all   *viper.Viper
	dbMap map[string]*databaseConf
}

// database 数据库
type databaseConf struct {
	User        string `json:"user"`
	Password    string `json:"password"`
	Host        string `json:"host"`
	Name        string `json:"name"`
	TablePrefix string `json:"tablePrefix"`
	Port        int    `json:"port"`
}
type serverConf struct {
	Addr string `json:"addr"`
}

var confs *conf

func init() {
	confs = &conf{}
	confs.initConfig()
}

func Conf() *conf {
	return confs
}

func (c *conf) initConfig() {
	fileName := "config"
	fileType := "yaml"
	filePath := "."
	viper.SetConfigName(fileName)
	viper.SetConfigType(fileType)
	viper.AddConfigPath(filePath)
	err1 := viper.ReadInConfig()

	//读取区分环境变量的
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "local"
	}
	viper.SetConfigName(fmt.Sprintf("%s.%s", fileName, env))
	viper.SetConfigType(fileType)
	viper.AddConfigPath(filePath)
	err2 := viper.MergeInConfig()
	if err1 != nil && err2 != nil {
		log.Fatalln(err1, err2)
	}
	c.all = viper.GetViper()
	c.initDBConfig()
}
func (c *conf) initDBConfig() {
	dbs := c.all.GetStringMap("datasource")
	c.dbMap = make(map[string]*databaseConf)
	for k, v := range dbs {
		db := &databaseConf{}
		jsonMap, err := json.Marshal(v)
		if err != nil {
			log.Fatalln(err)
		}
		err = json.Unmarshal([]byte(jsonMap), db)
		if err != nil {
			log.Fatalln(err)
		}
		c.dbMap[k] = db
	}
}
func (c *conf) GetViper() *viper.Viper {
	return c.all
}
func (c *conf) GetString(key string) string {
	return c.all.GetString(key)
}
func (c *conf) GetInt(key string) int {
	return c.all.GetInt(key)
}
func (c *conf) GetAllDB() map[string]*databaseConf {
	return c.dbMap
}
func (c *conf) GetDB(name string) *databaseConf {
	if name == "" {
		name = "default"
	}
	if db, ok := c.dbMap[name]; ok {
		return db
	}
	return nil
}
func (c *conf) GetServer() *serverConf {
	ser := &serverConf{}
	if err := c.all.UnmarshalKey("server", ser); err != nil || ser.Addr == "" {
		ser.Addr = ":8080"
	}
	return ser
}
