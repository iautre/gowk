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
					log.Errorf(c, (err.(runtime.Error)).Error())
					Response().Fail(c, ERR_UN, nil)
					// c.Abort()
				default: // 非运行时错误
					log.Errorf(c, fmt.Sprintf("%v", err))
					Response().Fail(c, ERR_UN, nil)
				}
			}
		}()
		c.Next()
	}
}
