package gowk

// goroutinePool 通过 semaphore channel 限制并发 goroutine 数量。
// Go() 始终异步执行，不会阻塞调用方。
type goroutinePool struct {
	sem chan struct{}
}

var goroutines = &goroutinePool{
	sem: make(chan struct{}, 1000),
}

// SetGoroutinePoolSize 在程序启动时调用，设置最大并发 goroutine 数（默认 1000）。
func SetGoroutinePoolSize(n int) {
	goroutines.sem = make(chan struct{}, n)
}

// Go 异步执行 f，在并发数达到上限时阻塞等待空闲槽位，但不阻塞调用方 goroutine。
func Go(f func()) {
	go func() {
		goroutines.sem <- struct{}{} // 占用一个槽位
		defer func() { <-goroutines.sem }()
		f()
	}()
}
