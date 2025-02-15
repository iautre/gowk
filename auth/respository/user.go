package respository

import (
	"context"
	"github.com/iautre/gowk/auth/model"
	"github.com/iautre/gowk/auth/respository/gorm"
)

type UserRepository interface {
	Save(ctx context.Context, m *model.User) error
	GetById(ctx context.Context, id int64) (*model.User, error)
	GetByToken(ctx context.Context, token string) (*model.User, error)
	GetByPhone(ctx context.Context, phone string) (*model.User, error)
}

func NewUserRepository(db ...string) UserRepository {
	return &gorm.GormUser{}
}
