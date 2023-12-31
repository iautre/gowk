package gowk

import (
	"sync"

	"github.com/iautre/gowk/conf"
)

// var once sync.Once

func init() {
	var wg sync.WaitGroup

	if conf.DB != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gormDBs.Init("", conf.DB, false)
		}()
	}
	if conf.Mongo != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mongos.Init("", conf.Mongo, false)
		}()
	}
	wg.Wait()
}
