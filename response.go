package gowk

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type errCode struct {
	ErrCode       int
	ErrMsg        string
	ERR_NOTFOUND  *errCode
	ERR_OK        *errCode
	ERR_SERVERERR *errCode
	ERR_NOSERVER  *errCode
	ERR_DBERR     *errCode
	ERR           *errCode
	ERR_RESERR    *errCode
}

// 错误码
// 11-- token类
// 12-- 用户类
// 21--- coding类
// 22--- imoxin类型
var (
	response     *errCode
	responseOnce sync.Once
)

func initErr() {
	response = &errCode{}
	response.ERR_NOTFOUND = response.initError(404, "未找到")
	response.ERR_OK = response.initError(0, "请求成功")
	response.ERR_SERVERERR = response.initError(98, "服务异常")
	response.ERR_NOSERVER = response.initError(99, "服务不存在")
	response.ERR_DBERR = response.initError(21, "查询失败")
	response.ERR = response.initError(-1, "请求异常")
	response.ERR_RESERR = response.initError(9, "返回异常")
}

func Response() *errCode {
	if response == nil {
		responseOnce.Do(initErr)
	}
	return response
}

func (e *errCode) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, e.responseToMap(e.ERR_OK, data))
}
func (e *errCode) Fail(c *gin.Context, code *errCode, err error) {
	c.JSON(http.StatusOK, e.responseToMap(code, nil))
}

func (e *errCode) initError(code int, msg string) *errCode {
	return &errCode{
		ErrCode: code,
		ErrMsg:  msg,
	}
}

func (e *errCode) NewError(msg string, err ...interface{}) *errCode {
	//log.Error("这是错误", err...)
	e.ERR.ErrMsg = msg
	return e.ERR
}

func (e *errCode) responseToMap(errcode *errCode, data interface{}) map[string]interface{} {

	res := make(map[string]interface{})
	if data == nil {
		e.errToMap(errcode, res)
		return res
	}
	jsonMap, err := json.Marshal(data)
	if err != nil {
		e.errToMap(e.ERR_RESERR, res)
		return res
	}
	err = json.Unmarshal([]byte(jsonMap), &res)
	if err != nil {
		e.errToMap(e.ERR_RESERR, res)
		return res
	}
	e.errToMap(errcode, res)
	return res
}

func (e *errCode) errToMap(errcode *errCode, res map[string]interface{}) {
	res["errcode"] = errcode.ErrCode
	res["errmsg"] = errcode.ErrMsg
}
