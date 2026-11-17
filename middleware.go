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
			myErr, ok := errors.AsType[*ErrorCode](ctx.Errors.Last())
			if !ok {
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
		// 无错误、未写入响应、且状态码仍是默认 200 时才补默认 OK。
		// 只看 Written() 不够：它判断的是"是否写出过字节"，而 ctx.Redirect(302)
		// 只登记状态码不落盘，POST 下 http.Redirect 也不写 body，此时 Written()
		// 为 false 但状态码已是 302，补 JSON 会把重定向盖成 200。
		if !ctx.Writer.Written() && ctx.Writer.Status() == http.StatusOK {
			ctx.JSON(http.StatusOK, OK)
			ctx.Abort()
		}
	}
}
