package gowk

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type err struct {
	errCode int
	errMsg  string
}

var _ Error = (*err)(nil)

type Error interface {
	Message(code *err, data interface{}) map[string]interface{}
	Success(c *gin.Context, data interface{})
	Fail(c *gin.Context, code *err, err error)
}

// 错误码
// 11-- token类
// 12-- app类
// 13-- 用户类
// 14 -- 公共

// 21--- coding类
// 22--- imoxin类型

// 5-- socket类

var (
	response = &err{}
	OK       = response.initError(0, "成功")
	ERR      = response.initError(-1, "失败")

	ERR_TOKEN = response.initError(1101, "无效token")

	ERR_NOAPP = response.initError(12404, "无效app")

	ERR_PARAM    = response.initError(1401, "参数错误")
	ERR_NOTFOUND = response.initError(404, "未找到")
	ERR_SERVER   = response.initError(98, "服务异常")
	ERR_NOSERVER = response.initError(99, "服务不存在")
	ERR_DBERR    = response.initError(21, "查询失败")

	ERR_RESERR     = response.initError(9, "返回异常")
	ERR_WS_CONTENT = response.initError(0, "已连接")
	ERR_WS_CLOSE   = response.initError(0, "已连接")
)

func Response() Error {
	return response
}

func (e *err) Message(code *err, data interface{}) map[string]interface{} {
	return e.responseToMap(code, data)
}

func (e *err) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, e.responseToMap(OK, data))
	c.Abort()
}
func (e *err) Fail(c *gin.Context, code *err, err error) {
	Log().Error(c, err.Error())
	c.JSON(http.StatusOK, e.responseToMap(code, nil))
	c.Abort()
}

func (e *err) initError(code int, msg string) *err {
	return &err{
		errCode: code,
		errMsg:  msg,
	}
}

func (e *err) NewError(msg string, errs ...interface{}) *err {
	//log.Error("这是错误", err...)
	return &err{
		errCode: 0,
		errMsg:  msg,
	}
}

func (e *err) responseToMap(errcode *err, data interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	if data == nil {
		e.errToMap(errcode, res)
		return res
	}
	jsonMap, err := json.Marshal(data)
	if err != nil {
		e.errToMap(ERR_RESERR, res)
		return res
	}
	err = json.Unmarshal([]byte(jsonMap), &res)
	if err != nil {
		e.errToMap(ERR_RESERR, res)
		return res
	}
	e.errToMap(errcode, res)
	return res
}

func (e *err) errToMap(errcode *err, res map[string]interface{}) {
	res["errcode"] = errcode.errCode
	res["errmsg"] = errcode.errMsg
}
