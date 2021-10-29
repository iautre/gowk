package gowk

import (
	"sync"

	"github.com/gin-gonic/gin"
)

type middleware struct{}

var (
	middlewares    *middleware
	middlewareOnce sync.Once
)

func Middleware() *middleware {
	if middlewares == nil {
		middlewareOnce.Do(func() {
			middlewares = &middleware{}
		})
	}
	return middlewares
}

func (m *middleware) RequestLog() gin.HandlerFunc {
	r := &requestLog{}
	return func(c *gin.Context) {
		r.RequestInLog(c)
		defer r.RequestOutLog(c)
		c.Next()
	}
}

type requestLog struct {
}

// 请求进入日志
func (r *requestLog) RequestInLog(c *gin.Context) {

	if traceid := c.Request.Header.Get("traceid"); traceid == "" {
		traceid = UUID()
		c.Set("traceid", traceid)
		c.Request.Header.Set("traceid", traceid)
	}
	if spanId := c.Request.Header.Get("spanid"); spanId != "" {
		c.Set("pspanid", spanId)
		c.Request.Header.Set("pspanid", spanId)
	}
	spanId := UUID()
	c.Set("spanId", spanId)
	c.Request.Header.Set("spanId", spanId)

}

// 请求输出日志
func (r *requestLog) RequestOutLog(c *gin.Context) {
	// after request
	//endTime := time.Now()

	//trace, _ := c.Get("trace")
	//traceContext := trace.((*log.TraceContext))

	//log.Printf(" [INFO] %s %s  %s\n", endTime, traceContext.TraceId, traceContext.TraceId)
}
