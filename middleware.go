package gowk

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"runtime"
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
					// 返回错误信息
					end(c, e)
				case runtime.Error: // 运行时错误
					e := err.(runtime.Error)
					slog.ErrorContext(c, e.Error())
					end(c, Error(e))
				default: // 非运行时错误
					slog.ErrorContext(c, fmt.Sprintf("recover type: %s", tp))
					slog.ErrorContext(c, fmt.Sprintf("%v", err))
					end(c, NewError(fmt.Sprintf("%v", err)))
				}
			}
		}()
		c.Next()
	}
}

// LogTrace 全局日志链路
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
			if !errors.As(ctx.Errors.Last(), &myErr) {
				myErr = NewError(ctx.Errors.Last().Err.Error())
			}
			if myErr.Status != 0 {
				ctx.JSON(myErr.Status, myErr)
			} else {
				ctx.JSON(http.StatusOK, myErr)
			}
			ctx.Abort()
			return
		}
		if ctx.IsAborted() == false {
			ctx.JSON(http.StatusOK, OK)
			ctx.Abort()
		}
	}
}
