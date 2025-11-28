package gowk

import (
	"github.com/gin-gonic/gin"
)

type Service[T any] struct {
	Ctx *gin.Context
}

func NewService[T any](ctx *gin.Context) *Service[T] {
	return &Service[T]{
		Ctx: ctx,
	}
}
func (s *Service[T]) Page(pageModel *PageModel[T], queryParam *T) (*PageModel[T], error) {
	//var t T
	//var resStructList []*T
	//if err := DB(s.Ctx).Query(s.Ctx).Where(queryParam).Count(&pageModel.Total).Error; err != nil {
	//	return nil, err
	//}
	//if err := DB(s.Ctx).Model(&t).Where(queryParam).Scopes(Paginate(pageModel)).Find(&resStructList).Error; err != nil {
	//	return nil, err
	//}
	//pageModel.Records = resStructList
	return pageModel, nil
}

func (s *Service[T]) One(queryParam *T) (T, error) {
	var model T
	//err := DB(s.Ctx).Where(queryParam).First(&model).Error
	return model, nil
}
func (s *Service[T]) Update(postParam *T) error {
	//return DB(s.Ctx).Updates(postParam).Error
	return nil
}
func (s *Service[T]) Save(postParam *T) error {
	//return DB(s.Ctx).Save(postParam).Error
	return nil
}
