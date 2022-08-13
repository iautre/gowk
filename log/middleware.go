package log

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	StartTime string = "startTime"
)

func RequestMiddleware() gin.HandlerFunc {
	r := &requestLog{}
	return func(c *gin.Context) {
		r.RequestInLog(c)
		defer r.RequestOutLog(c)
		c.Next()
	}
}

type requestLog struct{}

// 请求进入日志
func (r *requestLog) RequestInLog(c *gin.Context) {
	startTime := time.Now()
	c.Set(StartTime, startTime)
	if traceId := c.Request.Header.Get(TraceId); traceId == "" {
		traceId = uuid.NewString()
		c.Set(TraceId, traceId)
		c.Request.Header.Set(TraceId, traceId)
	}
	if spanId := c.Request.Header.Get(SpanId); spanId != "" {
		c.Set(PspanId, spanId)
		c.Request.Header.Set(PspanId, spanId)
	}
	spanId := uuid.NewString()
	c.Set(SpanId, spanId)
	c.Request.Header.Set(SpanId, spanId)
	// msg := fmt.Sprintf("%s %s %s Header: %v Body: %v start", c.ClientIP(), c.Request.Method, c.Request.RequestURI, c.Request.Header, c.Request.Body)
	msg := &H{
		"type":   "start",
		"ip":     c.ClientIP(),
		"method": c.Request.Method,
		"uri":    c.Request.RequestURI,
		"header": c.Request.Header,
		"body":   c.Request.Body,
	}
	std.Trace(c, msg)
}

// 请求输出日志
func (r *requestLog) RequestOutLog(c *gin.Context) {
	// after request
	endTime := time.Now()
	startTime, _ := c.Get(StartTime)
	usedTime := endTime.Sub(startTime.(time.Time)).Milliseconds()
	msg := fmt.Sprintf("%d end %dms", c.Writer.Status(), usedTime)
	std.Trace(c, msg)
}
