package gowk

import (
	"github.com/gin-gonic/gin"
)

var _ Error = (*errorCode)(nil)

type Error interface {
	i()
	Message(code *errorCode, data any) []byte
	Success(c *gin.Context, data any)
	Fail(c *gin.Context, code *errorCode, err error)
}

var response = &errorCode{}

func Response() Error {
	return response
}
