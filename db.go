package gowk

import (
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type dbType interface {
	*gorm.DB | *mongo.Client
}

func DB[T dbType](names ...string) (t T) {
	var ti interface{} = t
	switch ti.(type) {
	case *gorm.DB:
		ti = Mysql()
	case *mongo.Client:
		ti = Mongo()
	}
	return ti.(T)
}
