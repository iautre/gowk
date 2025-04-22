package repository

import (
	"context"
	"github.com/iautre/gowk/auth/model"
	"github.com/iautre/gowk/auth/repository/gorm"
)

type BaseInterface interface {
	AutoMigrate(ctx context.Context, a ...any) error
}

func NewBaseRepository() BaseInterface {
	var p gorm.GormBase
	return &p
}

func init() {
	NewBaseRepository().AutoMigrate(context.TODO(), &model.User{}, &model.App{}, &model.AppData{})
}
