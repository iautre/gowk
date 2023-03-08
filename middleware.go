package gowk

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk/log"
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

// 全局统一处理错误
func (m *middleware) Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				switch tp := err.(type) {
				case *ErrorCode: // 自定义异常
					e := err.(*ErrorCode)
					log.Error(c, e.Msg, e.err)
					// 返回错误信息
					Response().Fail(c, e, e.err)
					// c.Abort()
				case runtime.Error: // 运行时错误
					e := err.(runtime.Error)
					log.Error(c, e.Error(), e)
					Response().Fail(c, ERR, nil)
					// c.Abort()
				default: // 非运行时错误
					log.Error(c, fmt.Sprintf("recover type: %s", tp), nil)
					log.Error(c, fmt.Sprintf("%v", err), nil)
					Response().Fail(c, ERR, nil)
				}
			}
		}()
		c.Next()
	}
}

// 返回http原生状态码
func (m *middleware) Recover2() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				switch tp := err.(type) {
				case *ErrorCode: // 自定义异常
					e := err.(*ErrorCode)
					log.Error(c, e.Msg, e.err)
					c.String(e.Code, e.Msg)
				case runtime.Error: // 运行时错误
					e := err.(runtime.Error)
					log.Error(c, e.Error(), e)
					c.String(http.StatusInternalServerError, err.(runtime.Error).Error())
				default: // 非运行时错误
					log.Error(c, fmt.Sprintf("recover type: %s", tp), nil)
					log.Error(c, fmt.Sprintf("%v", err), nil)
					c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
