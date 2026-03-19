package gowk

import "github.com/gin-gonic/gin"

type Handler[T any] struct {
}

func NewHandler[T any]() *Handler[T] {
	return &Handler[T]{}
}

func (h *Handler[T]) Page() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var page PageModel[T]
		if err := ctx.ShouldBind(&page); err != nil {
			ctx.Error(err)
			return
		}
		var t T
		if err := ctx.ShouldBind(&t); err != nil {
			ctx.Error(err)
			return
		}
		service := NewService[T](ctx)
		res, err := service.Page(&page, &t)
		if err != nil {
			ctx.Error(err)
			return
		}
		Success(ctx, res)
	}
}

func (h *Handler[T]) Save() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var t T
		if err := ctx.ShouldBind(&t); err != nil {
			ctx.Error(err)
			return
		}
		service := NewService[T](ctx)
		if err := service.Save(&t); err != nil {
			ctx.Error(err)
			return
		}
		Success(ctx, t)
	}
}

func (h *Handler[T]) Update() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var t T
		if err := ctx.ShouldBind(&t); err != nil {
			ctx.Error(err)
			return
		}
		service := NewService[T](ctx)
		if err := service.Update(&t); err != nil {
			ctx.Error(err)
			return
		}
		Success(ctx, t)
	}
}

func (h *Handler[T]) One() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var t T
		if err := ctx.ShouldBind(&t); err != nil {
			ctx.Error(err)
			return
		}
		service := NewService[T](ctx)
		res, err := service.One(&t)
		if err != nil {
			ctx.Error(err)
			return
		}
		Success(ctx, res)
	}
}
