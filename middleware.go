package gowk

import (
	"encoding/json"
	"fmt"
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
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				switch err.(type) {
				case string: // 自定义异常
					str, _ := err.(string)
					var errMsg *ErrorCode
					if err := json.Unmarshal([]byte(str), &errMsg); err != nil {
						log.Error(ctx, err.Error(), err)
						Response().Fail(ctx, ERR_UN, err)
						// c.Abort()
						return
					}
					log.Error(ctx, errMsg.Msg, errMsg.err)
					// 返回错误信息
					Response().Fail(ctx, errMsg, nil)
					// c.Abort()
				case runtime.Error: // 运行时错误
					log.Error(ctx, (err.(runtime.Error)).Error(), err.(runtime.Error))
					Response().Fail(ctx, ERR_UN, nil)
					// c.Abort()
				default: // 非运行时错误
					log.Error(ctx, fmt.Sprintf("%v", err), err.(error))
					Response().Fail(ctx, ERR_UN, nil)
				}
			}
		}()
		ctx.Next()
	}
}
