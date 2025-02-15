package gorm

import (
	"context"
	"github.com/iautre/gowk/auth/model"

	"github.com/iautre/gowk"
)

type GormUser struct{}

func (r *GormUser) Save(ctx context.Context, m *model.User) error {
	return gowk.DB(ctx).Save(m).Error

}
func (r *GormUser) GetByKey(ctx context.Context, key string) (*model.User, error) {
	var d model.User
	tx := gowk.DB(ctx).Where("key = ?", key).First(&d)
	return &d, tx.Error
}
func (r *GormUser) GetById(ctx context.Context, id int64) (*model.User, error) {
	var d model.User
	tx := gowk.DB(ctx).First(&d, id)
	return &d, tx.Error
}
func (r *GormUser) GetByToken(ctx context.Context, token string) (*model.User, error) {
	var d model.User
	tx := gowk.DB(ctx).Where("token = ?", token).First(&d)
	return &d, tx.Error
}
func (r *GormUser) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	var d model.User
	tx := gowk.DB(ctx).Where("phone = ?", phone).First(&d)
	return &d, tx.Error
}
