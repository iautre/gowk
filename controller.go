package gowk

import (
	"github.com/gin-gonic/gin"
)

type Controller struct {
}

//分页查询
func (s *Controller) GetList(model interface{}) func(*gin.Context) {
	return func(c *gin.Context) {
		pageModel := &PageModel{}
		if err := c.ShouldBindQuery(pageModel); err != nil {
			Response().Fail(c, ERR, err)
		}
		queryParam := CopyToStruct(model)
		if err := c.ShouldBindQuery(queryParam); err != nil {
			Response().Fail(c, ERR, err)
		}
		service := NewService(c)
		res, err := service.GetList(pageModel, queryParam)
		if err != nil {
			Response().Fail(c, ERR, err)
			return
		}
		Response().Success(c, res)
	}
}

func (s *Controller) GetOne(model interface{}) func(*gin.Context) {
	return func(c *gin.Context) {
		queryParam := CopyToStruct(model)
		if err := c.ShouldBindQuery(queryParam); err != nil {
			Response().Fail(c, ERR, err)
		}
		res, err := NewService(c).GetOne(queryParam)
		if err != nil {
			Response().Fail(c, ERR, err)
			return
		}
		Response().Success(c, res)
	}
}
func (s *Controller) Update(model interface{}) func(*gin.Context) {
	return func(c *gin.Context) {
		postParam := CopyToStruct(model)
		if err := c.ShouldBind(postParam); err != nil {
			Response().Fail(c, ERR, err)
		}
		res, err := NewService(c).Update(postParam)
		if err != nil {
			Response().Fail(c, ERR, err)
			return
		}
		Response().Success(c, res)
	}
}
func (s *Controller) Save(model interface{}) func(*gin.Context) {
	return func(c *gin.Context) {
		postParam := CopyToStruct(model)
		if err := c.ShouldBind(postParam); err != nil {
			Response().Fail(c, ERR, err)
		}
		res, err := NewService(c).Save(postParam)
		if err != nil {
			Response().Fail(c, ERR, err)
			return
		}
		Response().Success(c, res)
	}
}
