package gowk

import (
	"fmt"
	"sync"
	"time"

	gromMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type mysql struct {
	dbs map[string]*gorm.DB
}

var mysqls *mysql

var dbOnce sync.Once

func DB(names ...string) *gorm.DB {
	if mysqls == nil {
		dbOnce.Do(func() {
			mysqls = &mysql{}
			mysqls.initAllDB()
		})
	}
	if len(mysqls.dbs) == 0 {
		return nil
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
	return nil
}

func (m *mysql) initAllDB() {
	m.dbs = make(map[string]*gorm.DB)
	dbConfs := Conf().GetAllDB()
	var wg sync.WaitGroup
	wg.Add(len(dbConfs))
	for key, dbConf := range dbConfs {
		go func(m *mysql, key string, dbConf *databaseConf) {
			defer wg.Done()
			dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
				dbConf.User,
				dbConf.Password,
				dbConf.Host,
				dbConf.Port,
				dbConf.Name)
			m.dbs[key] = m.initDB(dsn)
		}(m, key, dbConf)
	}
	wg.Wait()
}
func (m *mysql) initDB(dsn string) *gorm.DB {
	gdb, err := gorm.Open(gromMysql.Open(dsn), &gorm.Config{
		Logger: Log().GromLogger(),
		NamingStrategy: schema.NamingStrategy{
			//TablePrefix: "gormv2_",
			SingularTable: true,
		},
	})

	//db.SetLogger(util.Log())
	if err != nil {
		fmt.Println(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		fmt.Println(err)
	}
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
	return gdb
}
