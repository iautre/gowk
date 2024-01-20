package gowk

import (
	"fmt"
	"log"
	"time"

	"github.com/iautre/gowk/conf"
	gromMysql "gorm.io/driver/mysql"
	gromPostgresql "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type gormDB struct {
	dbs map[string]*gorm.DB
}

var gormDBs *gormDB

func GormDB(names ...string) *gorm.DB {
	if gormDBs == nil {
		log.Panic("未配置数据库")
	}
	if len(names) == 0 {
		if db, ok := gormDBs.dbs["default"]; ok {
			return db
		}
		for _, v := range gormDBs.dbs {
			return v
		}
	}
	if db, ok := gormDBs.dbs[names[0]]; ok {
		return db
	}
	log.Panic("未找到配置数据库")
	return nil
}

func initGormDBs(dbConf *conf.DatabaseConf, reset bool) {
	if gormDBs == nil {
		gormDBs = &gormDB{}
	}
	if gormDBs.dbs == nil {
		gormDBs.dbs = make(map[string]*gorm.DB)
	}
	if _, ok := gormDBs.dbs[dbConf.Key]; !ok || reset {
		gormDBs.dbs[dbConf.Key] = initGormDB(dbConf)
	}
}

func initGormDB(dbConf *conf.DatabaseConf) *gorm.DB {
	var dialector gorm.Dialector
	if dbConf.Type == "mysql" {
		dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
			dbConf.User,
			dbConf.Password,
			dbConf.Host,
			dbConf.Port,
			dbConf.Name)
		dialector = gromMysql.Open(dsn)
	}
	if dbConf.Type == "postgresql" || dbConf.Type == "postgres" {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			dbConf.Host,
			dbConf.User,
			dbConf.Password,
			dbConf.Name,
			dbConf.Port,
		)
		dialector = gromPostgresql.Open(dsn)
	}
	if dialector == nil {
		log.Panic("未找到配置数据库")
	}
	gdb, err := gorm.Open(dialector, &gorm.Config{
		Logger: &GromLogger{},
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   dbConf.TablePrefix,
			SingularTable: true,
		},
	})
	if err != nil {
		log.Panic(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		log.Panic(err)
	}
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
	return gdb
}
