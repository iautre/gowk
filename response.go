package gowk

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Response(ctx *gin.Context, statusCode int, data any, err error) {
	if http.StatusOK == statusCode {
		ctx.JSON(statusCode, data)
	} else {
		ctx.JSON(statusCode, gin.H{
			"error": err.Error(),
			"code":  statusCode,
		})
	}
	ctx.Abort()
}

func Success(c *gin.Context, data any) {
	res := &ErrorCode{
		Code: OK.Code,
		Msg:  OK.Msg,
		Data: data,
	}
	end(c, res)
}
func Fail(c *gin.Context, code *ErrorCode) {
	end(c, code)
}
func end(c *gin.Context, code *ErrorCode) {
	if c.IsAborted() {
		return
	}
	slog.InfoContext(c, fmt.Sprintf("result: %v", code.String()))
	if code.Status != 0 {
		c.JSON(code.Status, code)
	} else {
		c.JSON(http.StatusOK, code)
	}
	c.Abort()
}
