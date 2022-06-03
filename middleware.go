package gowk

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

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
	startTime := time.Now()
	c.Set("startTime", startTime)
	if traceId := c.Request.Header.Get("traceId"); traceId == "" {
		traceId = UUID()
		c.Set("traceId", traceId)
		c.Request.Header.Set("traceId", traceId)
	}
	if spanId := c.Request.Header.Get("spanId"); spanId != "" {
		c.Set("pspanId", spanId)
		c.Request.Header.Set("pspanId", spanId)
	}
	spanId := UUID()
	c.Set("spanId", spanId)
	c.Request.Header.Set("spanId", spanId)
	msg := fmt.Sprintf("%s %s %s start", c.ClientIP(), c.Request.Method, c.Request.RequestURI)
	Log().Trace(c, msg)
}

// 请求输出日志
func (r *requestLog) RequestOutLog(c *gin.Context) {
	// after request
	endTime := time.Now()
	startTime, _ := c.Get("startTime")
	usedTime := endTime.Sub(startTime.(time.Time)).Milliseconds()
	msg := fmt.Sprintf("%d end %dms", c.Writer.Status(), usedTime)
	Log().Trace(c, msg)
}

//	全局统一处理错误
func (m *middleware) Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				switch err.(type) {
				case string: // 自定义异常
					str, _ := err.(string)
					var errMsg *ErrorCode
					if err := json.Unmarshal([]byte(str), &errMsg); err != nil {
						Response().Fail(c, ERR_UN, err)
						// c.Abort()
						return
					}
					// 返回错误信息
					Response().Fail(c, errMsg, nil)
					// c.Abort()
				case runtime.Error: // 运行时错误
					Log().Error(c, (err.(runtime.Error)).Error())
					Response().Fail(c, ERR_UN, nil)
					// c.Abort()
				default: // 非运行时错误
					Log().Error(c, fmt.Sprintf("%v", err))
					Response().Fail(c, ERR_UN, nil)
				}
			}
		}()
		c.Next()
	}
}
