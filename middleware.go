package gowk

import (
	"errors"
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
					e := tp
					if e.Code == OK.Code {
						return
					}
					// 返回错误信息
					end(c, e)
				case runtime.Error: // 运行时错误
					slog.ErrorContext(c, tp.Error())
					end(c, Error(tp))
				default: // 非运行时错误
					slog.ErrorContext(c, "recover", "type", fmt.Sprintf("%T", err), "value", err)
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
		ctx.JSON(http.StatusNotFound, NOT_FOUND)
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
			if !ctx.IsAborted() {
				if myErr.Status != 0 {
					ctx.JSON(myErr.Status, myErr)
				} else {
					ctx.JSON(http.StatusOK, myErr)
				}
				ctx.Abort()
			}
			return
		}
		if !ctx.IsAborted() {
			ctx.JSON(http.StatusOK, OK)
			ctx.Abort()
		}
	}
}
