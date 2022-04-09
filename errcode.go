package gowk

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type errorCode struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func Panic(e *errorCode) {
	data, err := json.Marshal(e)
	if err != nil {
		log.Panic(err)
	}
	log.Panic(string(data))
}
func NewErrorCode(code int, msg string) *errorCode {
	return &errorCode{
		Code: code,
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

func (e *errorCode) i() {}

func (e *errorCode) Message(code *errorCode, data any) []byte {
	res := &errorCode{
		Code: code.Code,
		Msg:  code.Msg,
		Data: data,
	}
	jsonByte, _ := json.Marshal(res)
	return jsonByte
}

// 成功消息
func (e *errorCode) Success(c *gin.Context, data any) {
	res := &errorCode{
		Code: OK.Code,
		Msg:  OK.Msg,
		Data: data,
	}
	c.JSON(http.StatusOK, res)
	c.Abort()
}

//失败消息
func (e *errorCode) Fail(c *gin.Context, code *errorCode, err error) {
	if err != nil {
		Log().Error(c, err.Error())
	}
	res := &errorCode{
		Code: code.Code,
		Msg:  code.Msg,
	}
	c.JSON(http.StatusOK, res)
	c.Abort()
}
