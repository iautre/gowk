package gowk

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Response(ctx *gin.Context, statusCode int, data any, err error) {
	if statusCode == http.StatusOK {
		ctx.JSON(statusCode, Result(data))
	} else {
		if err == nil {
			err = errors.New(strconv.Itoa(statusCode))
		}
		var ec *ErrorCode
		if errors.As(err, &ec) {
			ec.Status = statusCode
			ctx.JSON(statusCode, ec)
		} else {
			ctx.JSON(statusCode, &ErrorCode{
				Status: statusCode,
				Code:   statusCode,
				Msg:    err.Error(),
			})
		}
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
