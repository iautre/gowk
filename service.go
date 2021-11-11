package gowk

import "github.com/gin-gonic/gin"

type Service struct {
	Ctx *gin.Context
}

func NewService(ctx *gin.Context) *Service {
	return &Service{
		Ctx: ctx,
	}
}
func (s *Service) GetList(pageModel *PageModel, queryParam interface{}) (*PageModel, error) {
	model := CopyToStruct(queryParam)
	if pageModel.Page <= 0 {
		pageModel.Page = 1
	}
	if pageModel.Size <= 0 {
		pageModel.Size = 10
	}
	//newModel := CopyToStruct(model)
	resStructList := CopyToStructSlice(queryParam)
	if res := DB().WithContext(s.Ctx).Model(model).Where(queryParam).Limit(pageModel.Size).Offset((pageModel.Page - 1) * pageModel.Size).Count(&pageModel.Total); res.Error != nil {
		return nil, res.Error
	}
	if res := DB().WithContext(s.Ctx).Model(model).Where(queryParam).Limit(pageModel.Size).Offset((pageModel.Page - 1) * pageModel.Size).Find(resStructList); res.Error != nil {
		return nil, res.Error
	}
	pageModel.List = resStructList
	return pageModel, nil
}

func (s *Service) GetOne(queryParam interface{}) (interface{}, error) {
	model := CopyToStruct(queryParam)
	resStruct := CopyToStruct(queryParam)
	if res := DB().WithContext(s.Ctx).Model(model).Where(queryParam).First(resStruct); res.Error != nil {
		return nil, res.Error
	}
	return resStruct, nil
}
func (s *Service) Update(postParam interface{}) (interface{}, error) {
	model := CopyToStruct(postParam)
	if res := DB().WithContext(s.Ctx).Model(model).Updates(postParam); res.Error != nil {
		return nil, res.Error
	}
	return postParam, nil
}
func (s *Service) Save(postParam interface{}) (interface{}, error) {
	model := CopyToStruct(postParam)
	if res := DB().WithContext(s.Ctx).Model(model).Save(postParam); res.Error != nil {
		return nil, res.Error
	}
	return postParam, nil
}
