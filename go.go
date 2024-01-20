package gowk

import "sync"

type goroutine struct {
	maxNum int
	mutex  sync.Mutex
}

var goroutines = &goroutine{
	maxNum: 1000,
}

func Go(f func()) {
	if goroutines.maxNum > 0 {
		go goroutines.Exec(f)
	} else {
		f()
	}
}

func (g *goroutine) Start() {
	goroutines.mutex.Lock()
	defer goroutines.mutex.Unlock()
	goroutines.maxNum--
}
func (g *goroutine) Done() {
	goroutines.mutex.Lock()
	defer goroutines.mutex.Unlock()
	goroutines.maxNum++
}
func (g *goroutine) Exec(f func()) {
	g.Start()
	defer g.Done()
	f()
}
