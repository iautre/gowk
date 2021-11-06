package gowk

import (
	"github.com/gin-gonic/gin"
)

type Service struct {
}

//分页查询
func (s *Service) GetList(model interface{}) func(*gin.Context) {
	return func(c *gin.Context) {
		pageModel := &PageModel{}
		if err := c.ShouldBindQuery(pageModel); err != nil {
			Response().Fail(c, Response().ERR, err)
			return
		}
		if pageModel.Page <= 0 {
			pageModel.Page = 1
		}
		if pageModel.Size <= 0 {
			pageModel.Size = 10
		}
		queryParam := CopyToStruct(model)
		if err := c.ShouldBindQuery(queryParam); err != nil {
			Response().Fail(c, Response().ERR, err)
			return
		}
		//newModel := CopyToStruct(model)
		resStructList := CopyToStructSlice(model)
		if res := DB().WithContext(c).Model(model).Where(queryParam).Limit(pageModel.Size).Offset((pageModel.Page - 1) * pageModel.Size).Count(&pageModel.Total); res.Error != nil {
			Response().Fail(c, Response().ERR, res.Error)
			return
		}
		if res := DB().WithContext(c).Model(model).Where(queryParam).Limit(pageModel.Size).Offset((pageModel.Page - 1) * pageModel.Size).Find(resStructList); res.Error != nil {
			Response().Fail(c, Response().ERR, res.Error)
			return
		}
		pageModel.List = resStructList
		Response().Success(c, pageModel)
	}
}

func (s *Service) GetOne(model interface{}) func(*gin.Context) {
	return func(c *gin.Context) {
		queryParam := CopyToStruct(model)
		if err := c.ShouldBindQuery(queryParam); err != nil {
			Response().Fail(c, Response().ERR, err)
			return
		}
		resStruct := CopyToStruct(model)
		if res := DB().WithContext(c).Model(model).Where(queryParam).First(resStruct); res.Error != nil {
			Response().Fail(c, Response().ERR, res.Error)
			return
		}
		Response().Success(c, resStruct)
	}
}
func (s *Service) Update(model interface{}) func(*gin.Context) {
	return func(c *gin.Context) {
		postParam := CopyToStruct(model)
		if err := c.ShouldBind(postParam); err != nil {
			Response().Fail(c, Response().ERR, err)
			return
		}
		if res := DB().WithContext(c).Model(model).Updates(postParam); res.Error != nil {
			Response().Fail(c, Response().ERR, res.Error)
			return
		}
		Response().Success(c, postParam)
	}
}
func (s *Service) Save(model interface{}) func(*gin.Context) {
	return func(c *gin.Context) {
		postParam := CopyToStruct(model)
		if err := c.ShouldBind(postParam); err != nil {
			Response().Fail(c, Response().ERR, err)
			return
		}
		if res := DB().WithContext(c).Model(model).Save(postParam); res.Error != nil {
			Response().Fail(c, Response().ERR, res.Error)
			return
		}
		Response().Success(c, postParam)
	}
}
