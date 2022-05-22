package gowk

import (
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
		tmp = Mysql()
	case *mongo.Client:
		tmp = Mongo()
	}
	return tmp.(T)
}
func DB(names ...string) *gorm.DB {
	return Mysql(names...)
}
