package gowk

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				switch tp := err.(type) {
				case *ErrorCode:
					if tp.Code == OK.Code {
						return
					}
					end(c, tp)
				case runtime.Error:
					slog.ErrorContext(c, tp.Error())
					end(c, Error(tp))
				default:
					slog.ErrorContext(c, "recover", "type", fmt.Sprintf("%T", err), "value", err)
					end(c, NewError(fmt.Sprintf("%v", err)))
				}
			}
		}()
		c.Next()
	}
}

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
			// 只在响应未写入时才写入错误响应，避免污染流式/文件响应
			if !ctx.Writer.Written() {
				if myErr.Status != 0 {
					ctx.JSON(myErr.Status, myErr)
				} else {
					ctx.JSON(http.StatusOK, myErr)
				}
				ctx.Abort()
			}
			return
		}
		// 无错误且响应未写入时才补充默认 OK 响应
		if !ctx.Writer.Written() {
			ctx.JSON(http.StatusOK, OK)
			ctx.Abort()
		}
	}
}
