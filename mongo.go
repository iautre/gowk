package gowk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDB struct {
	dbs map[string]*mongo.Client
}

var mongoDBs *mongoDB

func initMongo() {
	dbConfs := Conf().GetAllDB("mongo")
	if len(dbConfs) > 0 {
		mongoDBs = &mongoDB{}
		mongoDBs.initAllDB(dbConfs)
	}
}

func Mongo(names ...string) *mongo.Client {
	if mongoDBs == nil {
		panic("未配置数据库")
	}
	if len(names) == 0 {
		if db, ok := mongoDBs.dbs["default"]; ok {
			return db
		}
		for _, v := range mongoDBs.dbs {
			return v
		}
	}
	if db, ok := mongoDBs.dbs[names[0]]; ok {
		return db
	}
	panic("未找到配置数据库")
}

func (m *mongoDB) initAllDB(dbConfs map[string]*databaseConf) {
	m.dbs = make(map[string]*mongo.Client)
	var wg sync.WaitGroup
	wg.Add(len(dbConfs))
	for key, dbConf := range dbConfs {
		go func(m *mongoDB, key string, dbConf *databaseConf) {
			defer wg.Done()
			dsn := fmt.Sprintf("mongodb://%s:%s@%s:%d",
				dbConf.User,
				dbConf.Password,
				dbConf.Host,
				dbConf.Port)
			m.dbs[key] = m.initDB(dsn, dbConf.MaxPoolSize)
		}(m, key, dbConf)
	}
	wg.Wait()
}

func (m *mongoDB) initDB(uri string, maxPoolSize uint64) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetMaxPoolSize(maxPoolSize)) // 连接池
	if err != nil {
		panic("mongo连接池异常")
		//Log().Error(err)
	}
	return client
}
