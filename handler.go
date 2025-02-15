package gowk

import "github.com/gin-gonic/gin"

type Handler[T any] struct {
}

func NewHandler[T any]() *Handler[T] {
	return &Handler[T]{}
}

func (c *Handler[T]) Page() gin.HandlerFunc {
	return func(c *gin.Context) {
		var page PageModel[T]
		if err := c.ShouldBind(&page); err != nil {
			Panic(ERR_PARAM, err)
		}
		var t T
		if err := c.ShouldBind(&t); err != nil {
			Panic(ERR_PARAM, err)
		}
		service := NewService[T](c)
		res, err := service.Page(&page, &t)
		if err != nil {
			Panic(ERR, err)
		}
		Success(c, res)
	}
}
