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
			c.Error(err)
			return
		}
		var t T
		if err := c.ShouldBind(&t); err != nil {
			c.Error(err)
			return
		}
		service := NewService[T](c)
		res, err := service.Page(&page, &t)
		if err != nil {
			c.Error(err)
			return
		}
		Success(c, res)
	}
}
func (c *Handler[T]) Add() gin.HandlerFunc {
	return func(c *gin.Context) {
		var t T
		if err := c.ShouldBind(&t); err != nil {
			c.Error(err)
			return
		}
		service := NewService[T](c)
		err := service.Save(&t)
		if err != nil {
			c.Error(err)
			return
		}
		Success(c, t)
	}
}
func (c *Handler[T]) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		var t T
		if err := c.ShouldBind(&t); err != nil {
			c.Error(err)
			return
		}
		service := NewService[T](c)
		err := service.Update(&t)
		if err != nil {
			c.Error(err)
			return
		}
		Success(c, t)
	}
}
func (c *Handler[T]) One() gin.HandlerFunc {
	return func(c *gin.Context) {
		var t T
		if err := c.ShouldBind(&t); err != nil {
			c.Error(err)
			return
		}
		service := NewService[T](c)
		res, err := service.One(&t)
		if err != nil {
			c.Error(err)
			return
		}
		Success(c, res)
	}
}
