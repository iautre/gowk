package gowk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iautre/gowk/conf"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDB struct {
	dbs map[string]*mongo.Client
	mu  sync.Mutex
}

var (
	mongos    *mongoDB
	mongoOnce sync.Once
)

func Mongo(names ...string) *mongo.Client {
	if mongos == nil {
		panic("未配置数据库")
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

func (m *mongoDB) Init(name string, dbConf *conf.MongoConf, reset bool) {
	if name == "" {
		name = "default"
	}
	if dbConf == nil {
		dbConf = conf.Mongo
	}
	if m == nil {
		m = &mongoDB{
			dbs: make(map[string]*mongo.Client),
		}
	}
	if _, ok := m.dbs[name]; !ok || reset {
		m.dbs[name] = m.initDB(dbConf)
	}
}

func (m *mongoDB) initDB(dbConf *conf.MongoConf) *mongo.Client {
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
