package gowk

import (
	"github.com/gin-gonic/gin"
)

// Service 提供泛型 CRUD 基础结构，方法需由业务层结合具体 DB 驱动实现。
// 框架目前使用 pgxpool（非 GORM），因此此处仅定义接口骨架，
// 业务层可通过嵌套 Service[T] 并重写方法来扩展。
type Service[T any] struct {
	Ctx *gin.Context
}

func NewService[T any](ctx *gin.Context) *Service[T] {
	return &Service[T]{Ctx: ctx}
}

// Page 列表分页查询，业务层须重写此方法以提供具体实现。
func (s *Service[T]) Page(pageModel *PageModel[T], queryParam *T) (*PageModel[T], error) {
	return pageModel, nil
}

// One 单条查询，业务层须重写此方法以提供具体实现。
func (s *Service[T]) One(queryParam *T) (T, error) {
	var model T
	return model, nil
}

// Update 更新操作，业务层须重写此方法以提供具体实现。
func (s *Service[T]) Update(postParam *T) error {
	return nil
}

// Save 新增操作，业务层须重写此方法以提供具体实现。
func (s *Service[T]) Save(postParam *T) error {
	return nil
}
