package gowk

import (
	"context"
	"fmt"
	"github.com/iautre/gowk/conf"
	gromMysql "gorm.io/driver/mysql"
	gromPostgresql "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
)

var default_gormDB *gorm.DB

func GormDB(ctx context.Context) *gorm.DB {
	if default_gormDB == nil {
		log.Panic("未配置数据库")
	}
	if tx, ok := ctx.Value(TRANSACTION).(*Transaction); ok && tx != nil {
		if tx.Tx != nil {
			return tx.Tx
		}
		if tx.Begin {
			tx.Tx = default_gormDB.WithContext(ctx).Begin()
			return tx.Tx
		}
	}
	return default_gormDB.WithContext(ctx)
}

func initGormDB(dbConf *conf.DatabaseConf) {
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
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Panic(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		log.Panic(err)
	}
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(dbConf.MaxIdleConns)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(dbConf.MaxOpenConns)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(dbConf.ConnMaxLifetime)
	default_gormDB = gdb
}
