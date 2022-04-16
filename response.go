package gowk

import (
	"github.com/gin-gonic/gin"
)

var _ Error = (*ErrorCode)(nil)

type Error interface {
	i()
	Message(code *ErrorCode, data any) []byte
	Success(c *gin.Context, data any)
	Fail(c *gin.Context, code *ErrorCode, err error)
}

var response = &ErrorCode{}

func Response() Error {
	return response
}
