package gowk

import (
	"sync"
	"time"
)

var (
	caches    *cache
	cacheOnce sync.Once
)

func Cache() *cache {
	if caches == nil {
		logOnce.Do(func() {
			caches = &cache{
				Data:     make(map[string]*item),
				Interval: 1,
			}
		})
	}
	return caches
}

type item struct {
	value    interface{}
	created  time.Time
	expireIn time.Duration
}

type cache struct {
	sync.RWMutex
	Data     map[string]*item
	Interval int
}

func (r *cache) Get(key string) interface{} {
	item := r.Data[key]
	if item.isExpire() {
		return nil
	}
	return item.value
}
func (r *cache) Set(key string, value interface{}, expireIns ...time.Duration) {
	r.Lock()
	var expireIn time.Duration
	expireIn = 0
	if len(expireIns) == 1 {
		expireIn = expireIns[0]
	}
	r.Data[key] = &item{
		value:    value,
		created:  time.Now(),
		expireIn: expireIn,
	}
	r.Unlock()
}

func (r *cache) Del(key string) {
	r.Lock()
	delete(r.Data, key)
	r.Unlock()
}

func (r *cache) GC() {

}

func (i *item) isExpire() bool {
	if i.expireIn == 0 {
		return false
	}
	return time.Now().Sub(i.created) > i.expireIn
}
