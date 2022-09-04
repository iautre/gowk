package gowk

import (
	"sync"

	"github.com/iautre/gowk/conf"
)

var once sync.Once

func init() {
	var wg sync.WaitGroup
	wg.Add(2)
	if conf.Mysql != nil {
		go func() {
			defer wg.Done()
			mysqls.Init("", conf.Mysql, false)
		}()
	}
	go func() {
		defer wg.Done()
		mongos.Init("", conf.Mongo, false)
	}()
	wg.Wait()
}
