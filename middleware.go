package gowk

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
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
					end(c, e)
					// c.Abort()
				case runtime.Error: // 运行时错误
					e := err.(runtime.Error)
					slog.ErrorContext(c, e.Error())
					end(c, NewError(e.Error()))
					// c.Abort()
				default: // 非运行时错误
					slog.ErrorContext(c, fmt.Sprintf("recover type: %s", tp))
					slog.ErrorContext(c, fmt.Sprintf("%v", err))
					end(c, NewError(fmt.Sprintf("%v", err)))
				}
			}
			end(c, OK)
		}()
		c.Next()
	}
}

// 全局日志链路
func LogTrace() gin.HandlerFunc {
	return RequestMiddleware()
}

func NotFound() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, NotFound)
		ctx.Abort()
	}
}

func GlobalErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		if len(ctx.Errors) > 0 {
			var myErr *ErrorCode
			if e, ok := ctx.Errors.Last().Err.(*ErrorCode); ok {
				myErr = e
			} else {
				myErr = NewError(ctx.Errors.Last().Err.Error())
			}
			ctx.JSON(http.StatusOK, myErr)
			ctx.Abort()
		}
	}
}
