package gowk

import (
	"encoding/json"
	"errors"
	"net/http"
)

type ErrorCode struct {
	Status int    `json:"-"`
	Code   int    `json:"code"`
	Msg    string `json:"error"`
	Data   any    `json:"data,omitempty"`
	err    error  `json:"-"`
	Sender string `json:"sender,omitempty"`
}

func Panic(e *ErrorCode, err ...error) {
	if len(err) > 0 {
		e.err = err[0]
	}
	panic(e)
}

func NewErrorCode(code int, msg string) *ErrorCode {
	return &ErrorCode{
		Code: code,
		Msg:  msg,
	}
}
func newAuthErrorCode(code int, msg string) *ErrorCode {
	return &ErrorCode{
		Status: http.StatusUnauthorized,
		Code:   code,
		Msg:    msg,
	}
}
func TokenError(msg string) *ErrorCode {
	return &ErrorCode{
		Status: http.StatusUnauthorized,
		Code:   http.StatusUnauthorized,
		Msg:    msg,
	}
}
func NewError(msg string) *ErrorCode {
	return &ErrorCode{
		Code: 99,
		Msg:  msg,
	}
}
func SocketMsg(data any, sender string) *ErrorCode {
	return &ErrorCode{
		Code:   OK.Code,
		Msg:    OK.Msg,
		Data:   data,
		Sender: sender,
	}
}
func Result(data any) *ErrorCode {
	return &ErrorCode{
		Code: OK.Code,
		Msg:  OK.Msg,
		Data: data,
	}
}
func Error(err error) *ErrorCode {
	var errs *ErrorCode
	if errors.As(err, &errs) {
		return errs
	}
	return &ErrorCode{
		Code: 99,
		Msg:  err.Error(),
		err:  err,
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
// 25 -- 数据库类

var (
	OK  = NewErrorCode(0, "成功")
	ERR = NewErrorCode(-1, "错误")

	NOT_FOUND = &ErrorCode{Status: 404, Code: 404, Msg: "未找到"}

	ERR_AUTH     = newAuthErrorCode(401, "认证失败")
	ERR_PARAM    = NewErrorCode(1401, "参数错误")
	ERR_SERVER   = NewErrorCode(98, "服务异常")
	ERR_NOSERVER = NewErrorCode(99, "服务不存在")
	ERR_DBERR    = NewErrorCode(2501, "查询失败")
	ERR_NODATA   = NewErrorCode(2502, "无数据")

	ERR_WS_CONTENT = NewErrorCode(0, "已连接")
	ERR_WS_CLOSE   = NewErrorCode(-1, "已连接")
)

// 实现error接口
func (e *ErrorCode) Error() string {
	return e.Msg
}

func (e *ErrorCode) String() string {
	jsonByte, _ := json.Marshal(e)
	return string(jsonByte)
}
