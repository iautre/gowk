package gowk

import (
	"sync"

	"github.com/iautre/gowk/conf"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type dbType interface {
	*gorm.DB | *mongo.Client
}

func DBs[T dbType](names ...string) (t T) {
	var tmp any = t
	switch tmp.(type) {
	case *gorm.DB:
		tmp = GormDB()
	case *mongo.Client:
		tmp = Mongo()
	}
	return tmp.(T)
}
func DB(names ...string) *gorm.DB {
	return GormDB(names...)
}

func initDB() {
	var wg sync.WaitGroup
	if conf.HasDB() {
		for _, v := range conf.DBs() {
			dbConf := v
			if v.Type == "mysql" || v.Type == "postgresql" || v.Type == "postgres" {
				wg.Add(1)
				go func() {
					defer wg.Done()
					initGormDBs(dbConf, false)
				}()
			} else if v.Type == "mongo" {
				wg.Add(1)
				go func() {
					defer wg.Done()
					mongos.Init(dbConf, false)
				}()
			}
		}

	}
	wg.Wait()
}
