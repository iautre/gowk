package gowk

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk/log"
)

type ErrorCode struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
	err  error  `json:"-"`
}

func Panic(e *ErrorCode, err error) {
	e.err = err
	data, _ := json.Marshal(e)
	panic(string(data))
}

func NewErrorCode(code int, msg string) *ErrorCode {
	return &ErrorCode{
		Code: code,
		Msg:  msg,
	}
}
func NewError(msg string) *ErrorCode {
	return &ErrorCode{
		Code: 99,
		Msg:  msg,
	}
}

// 错误码
// 1- 系统
// 2- 认证
// 3- socket
// 21-- token类
// 22-- app类
// 23-- 用户类
// 24 -- 公共

var (
	ERR    = NewErrorCode(-1, "错误")
	ERR_UN = NewErrorCode(-1, "未知错误")
	OK     = NewErrorCode(0, "成功")

	ERR_AUTH = NewErrorCode(401, "认证失败")

	ERR_TOKEN    = NewErrorCode(2101, "无效token")
	ERR_NOAPP    = NewErrorCode(12404, "无效app")
	ERR_PARAM    = NewErrorCode(1401, "参数错误")
	ERR_NOTFOUND = NewErrorCode(404, "未找到")
	ERR_SERVER   = NewErrorCode(98, "服务异常")
	ERR_NOSERVER = NewErrorCode(99, "服务不存在")
	ERR_DBERR    = NewErrorCode(21, "查询失败")

	ERR_NODATA = NewErrorCode(23, "无数据")

	ERR_RESERR = NewErrorCode(9, "返回异常")

	ERR_WS_CONTENT = NewErrorCode(0, "已连接")
	ERR_WS_CLOSE   = NewErrorCode(-1, "已连接")
)

func (e *ErrorCode) i() {}

func (e *ErrorCode) Message(code *ErrorCode, data any) []byte {
	res := &ErrorCode{
		Code: code.Code,
		Msg:  code.Msg,
		Data: data,
	}
	jsonByte, _ := json.Marshal(res)
	return jsonByte
}

// 成功消息
func (e *ErrorCode) Success(c *gin.Context, data any) {
	res := &ErrorCode{
		Code: OK.Code,
		Msg:  OK.Msg,
		Data: data,
	}
	log.Trace(c, fmt.Sprintf("result: %v", string(e.Message(res, data))), nil)
	c.JSON(http.StatusOK, res)
	c.Abort()
}

// 失败消息
func (e *ErrorCode) Fail(c *gin.Context, code *ErrorCode, err error) {
	if err != nil {
		log.Error(c, err.Error(), err)
	}
	res := &ErrorCode{
		Code: code.Code,
		Msg:  code.Msg,
	}
	log.Trace(c, fmt.Sprintf("result: %v", string(e.Message(res, nil))), nil)
	c.JSON(http.StatusOK, res)
	c.Abort()
}
