package gowk

import (
	"fmt"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk/log"
)

// 全局统一处理错误
func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				switch tp := err.(type) {
				case *ErrorCode: // 自定义异常
					e := err.(*ErrorCode)
					if e.Code == OK.Code {
						return
					}
					// log.Error(c, e.Msg, e.err)
					// 返回错误信息
					end(c, e, e.err)
					// c.Abort()
				case runtime.Error: // 运行时错误
					e := err.(runtime.Error)
					log.Error(c, e.Error(), e)
					end(c, ERR, nil)
					// c.Abort()
				default: // 非运行时错误
					log.Error(c, fmt.Sprintf("recover type: %s", tp), nil)
					log.Error(c, fmt.Sprintf("%v", err), nil)
					end(c, ERR, nil)
				}
			}
			end(c, OK)
		}()
		c.Next()
	}
}

// 全局日志链路
func LogTrace() gin.HandlerFunc {
	return log.RequestMiddleware()
}

func NotFound() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		Fail(ctx, NOT_FOUND)
	}
}
