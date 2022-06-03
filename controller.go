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
			Panic(ERR_PARAM)
		}
		var t T
		if err := c.ShouldBind(&t); err != nil {
			Panic(ERR_PARAM)
		}
		service := NewService[T](c)
		res, err := service.GetList(&page, &t)
		if err != nil {
			Panic(ERR)
		}
		Response().Success(c, res)
	}
}
