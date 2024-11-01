package gowk

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/iautre/gowk/conf"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	default_mongo *mongo.Client
)

func Mongo(ctx context.Context) *mongo.Client {
	if default_mongo == nil {
		log.Panic("未配置mongo数据库")
	}
	return default_mongo
}

func initMongo(dbConf *conf.DatabaseConf) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d",
		dbConf.User,
		dbConf.Password,
		dbConf.Host,
		dbConf.Port)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetMaxPoolSize(dbConf.MaxPoolSize)) // 连接池
	if err != nil {
		panic("mongo连接池异常")
		//Log().Error(err)
	}
	default_mongo = client
}
