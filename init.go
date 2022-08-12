package gowk

import "sync"

var once sync.Once

func init() {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		initMysql()
	}()
	go func() {
		defer wg.Done()
		initMongo()
	}()
	wg.Wait()
}
