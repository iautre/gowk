package gowk

import "github.com/gin-gonic/gin"

type Service[T any] struct {
	Ctx *gin.Context
}

func NewService[T any](ctx *gin.Context) *Service[T] {
	return &Service[T]{
		Ctx: ctx,
	}
}
func (s *Service[T]) GetList(pageModel *PageModel[T], queryParam *T) (*PageModel[T], error) {
	var t T
	var resStructList []*T
	if err := Mysql().WithContext(s.Ctx).Model(&t).Where(queryParam).Count(&pageModel.Total).Error; err != nil {
		return nil, err
	}
	if err := Mysql().WithContext(s.Ctx).Model(&t).Where(queryParam).Scopes(Paginate(pageModel)).Find(&resStructList).Error; err != nil {
		return nil, err
	}
	pageModel.List = resStructList
	return pageModel, nil
}

func (s *Service[T]) GetOne(queryParam *T) (T, error) {
	var model T
	err := Mysql().WithContext(s.Ctx).Where(queryParam).First(&model).Error
	return model, err
}
func (s *Service[T]) Update(postParam *T) error {
	return Mysql().WithContext(s.Ctx).Updates(postParam).Error
}
func (s *Service[T]) Save(postParam *T) error {
	return Mysql().WithContext(s.Ctx).Save(postParam).Error
}
