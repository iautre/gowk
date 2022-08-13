package gowk

import (
	"fmt"
	"sync"
	"time"

	"github.com/iautre/gowk/log"
	gromMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type mysql struct {
	dbs map[string]*gorm.DB
}

var mysqls *mysql

func initMysql() {
	dbConfs := Conf().GetAllDB("mysql")
	if len(dbConfs) > 0 {
		mysqls = &mysql{}
		mysqls.initAllDB(dbConfs)
	}
}

func Mysql(names ...string) *gorm.DB {
	if mysqls == nil {
		panic("未配置数据库")
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
	panic("未找到配置数据库")
}

func (m *mysql) initAllDB(dbConfs map[string]*databaseConf) {
	m.dbs = make(map[string]*gorm.DB)
	var wg sync.WaitGroup
	wg.Add(len(dbConfs))
	for key, dbConf := range dbConfs {
		go func(m *mysql, key string, dbConf *databaseConf) {
			defer wg.Done()
			m.dbs[key] = m.initDB(dbConf)
		}(m, key, dbConf)
	}
	wg.Wait()
}
func (m *mysql) initDB(dbConf *databaseConf) *gorm.DB {
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
		panic(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		panic(err)
	}
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
	return gdb
}
