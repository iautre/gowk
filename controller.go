package gowk

import "github.com/gin-gonic/gin"

type Controller[T any] struct {
}

func NewController[T any]() *Controller[T] {
	return &Controller[T]{}
}

func (c *Controller[T]) Page() gin.HandlerFunc {
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
