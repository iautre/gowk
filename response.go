package gowk

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, data any) {
	res := &ErrorCode{
		Code: OK.Code,
		Msg:  OK.Msg,
		Data: data,
	}
	end(c, res)
	Panic(res)
}
func Fail(c *gin.Context, code *ErrorCode) {
	end(c, code)
	Panic(code)
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
