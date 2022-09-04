package gowk

import (
	"fmt"
	logs "log"
	"time"

	"github.com/iautre/gowk/conf"
	"github.com/iautre/gowk/log"
	gromMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type mysqlDB struct {
	dbs map[string]*gorm.DB
}

var mysqls *mysqlDB

func Mysql(names ...string) *gorm.DB {
	if mysqls == nil {
		logs.Panic("未配置数据库")
	}
	if len(names) == 0 {
		if db, ok := mysqls.dbs["default"]; ok {
			return db
		}
		for _, v := range mysqls.dbs {
			return v
		}
	}
	if db, ok := mysqls.dbs[names[0]]; ok {
		return db
	}
	logs.Panic("未找到配置数据库")
	return nil
}

func (m *mysqlDB) Init(name string, dbConf *conf.MysqlConf, reset bool) {
	if name == "" {
		name = "default"
	}
	if dbConf == nil {
		dbConf = conf.Mysql
	}
	if m == nil {
		m = &mysqlDB{
			dbs: make(map[string]*gorm.DB),
		}
	}
	if _, ok := m.dbs[name]; !ok || reset {
		m.dbs[name] = m.initDB(dbConf)
	}
}

func (m *mysqlDB) initDB(dbConf *conf.MysqlConf) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		dbConf.User,
		dbConf.Password,
		dbConf.Host,
		dbConf.Port,
		dbConf.Name)
	gdb, err := gorm.Open(gromMysql.Open(dsn), &gorm.Config{
		Logger: &log.GromLogger{},
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   dbConf.TablePrefix,
			SingularTable: true,
		},
	})
	//db.SetLogger(util.Log())
	if err != nil {
		logs.Panic(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		logs.Panic(err)
	}
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
	return gdb
}
