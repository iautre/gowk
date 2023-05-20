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
	log.Trace(c, fmt.Sprintf("result: %v", string(Message(res, data))), nil)
	c.JSON(http.StatusOK, res)
	c.Abort()
}
func Fail(c *gin.Context, code *ErrorCode, err ...error) {
	if len(err) > 0 && err[0] != nil {
		log.Error(c, err[0].Error(), err[0])
	}
	res := &ErrorCode{
		Code: code.Code,
		Msg:  code.Msg,
	}
	log.Trace(c, fmt.Sprintf("result: %v", string(Message(res, nil))), nil)
	if code.Status != 0 {
		c.JSON(code.Status, res)
	} else {
		c.JSON(http.StatusOK, res)
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
