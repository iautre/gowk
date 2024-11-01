package gowk

import (
	"context"

	"github.com/iautre/gowk/conf"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type dbType interface {
	*gorm.DB | *mongo.Client
}

func Db[T dbType](ctx context.Context) (t T) {
	var tmp any = t
	switch tmp.(type) {
	case *gorm.DB:
		tmp = GormDB(ctx)
	case *mongo.Client:
		tmp = Mongo(ctx)
	}
	return tmp.(T)
}

func DB(ctx context.Context) *gorm.DB {
	return GormDB(ctx)
}

func initDB() {
	if conf.HasDB() {
		switch conf.DB().Type {
		case "mysql", "postgresql", "postgres":
			initGormDB(conf.DB())
		case "mongo":
			initMongo(conf.DB())
		}
	}
}
