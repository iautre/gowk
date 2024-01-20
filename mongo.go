package gowk

import (
	"context"
	"fmt"
	"time"

	"github.com/iautre/gowk/conf"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDB struct {
	dbs map[string]*mongo.Client
}

var (
	mongos *mongoDB = &mongoDB{}
)

func Mongo(names ...string) *mongo.Client {
	if mongos == nil || mongos.dbs == nil || len(mongos.dbs) == 0 {
		panic("未配置mongo数据库")
	}
	if len(names) == 0 {
		if db, ok := mongos.dbs["default"]; ok {
			return db
		}
		for _, v := range mongos.dbs {
			return v
		}
	}
	if db, ok := mongos.dbs[names[0]]; ok {
		return db
	}
	panic("未找到配置数据库")
}

func (m *mongoDB) Init(dbConf *conf.DatabaseConf, reset bool) {
	if dbConf.Key == "" {
		dbConf.Key = "default"
	}
	if m.dbs == nil {
		m.dbs = make(map[string]*mongo.Client)
	}
	if _, ok := m.dbs[dbConf.Key]; !ok || reset {
		m.dbs[dbConf.Key] = m.initDB(dbConf)
	}
}

func (m *mongoDB) initDB(dbConf *conf.DatabaseConf) *mongo.Client {
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
	return client
}
