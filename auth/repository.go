package auth

import (
	"context"

	"github.com/iautre/gowk"
)

type UserRepository interface {
	Save(ctx context.Context, m *User) error
	GetById(ctx context.Context, id int64) (*User, error)
	GetByToken(ctx context.Context, token string) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	UpdateToken(ctx context.Context, userId int64, token string) error
}

func NewUserRepository(db ...string) UserRepository {
	return &GormUser{}
}

type GormUser struct{}

func (r *GormUser) Save(ctx context.Context, m *User) error {
	return gowk.DB().Save(m).Error

}
func (r *GormUser) GetByKey(ctx context.Context, key string) (*User, error) {
	var d User
	tx := gowk.DB().Where("key = ?", key).First(&d)
	return &d, tx.Error
}
func (r *GormUser) GetById(ctx context.Context, id int64) (*User, error) {
	var d User
	tx := gowk.DB().First(&d, id)
	return &d, tx.Error
}
func (r *GormUser) GetByToken(ctx context.Context, token string) (*User, error) {
	var d User
	tx := gowk.DB().Where("token = ?", token).First(&d)
	return &d, tx.Error
}
func (r *GormUser) GetByPhone(ctx context.Context, phone string) (*User, error) {
	var d User
	tx := gowk.DB().Where("phone = ?", phone).First(&d)
	return &d, tx.Error
}
func (r *GormUser) UpdateToken(ctx context.Context, userId int64, token string) error {
	tx := gowk.DB().Model(&User{}).Where("id = ?", userId).Updates(map[string]interface{}{"token": token, "last_login": gowk.Now()})
	return tx.Error
}
