package gowk

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk/log"
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
func Fail(c *gin.Context, code *ErrorCode, err ...error) {
	end(c, code, err...)
	Panic(code, err...)
}
func end(c *gin.Context, code *ErrorCode, err ...error) {
	if c.IsAborted() {
		return
	}
	if len(err) > 0 && err[0] != nil {
		log.Error(c, err[0].Error(), err[0])
	}
	log.Trace(c, fmt.Sprintf("result: %v", string(Message(code, nil))), nil)
	if code.Status != 0 {
		c.JSON(code.Status, code)
	} else {
		c.JSON(http.StatusOK, code)
	}
	c.Abort()
}
func Message(code *ErrorCode, data any) []byte {
	res := &ErrorCode{
		Code: code.Code,
		Msg:  code.Msg,
		Data: data,
	}
	jsonByte, _ := json.Marshal(res)
	return jsonByte
}
