package gorm

import (
	"context"
	"github.com/iautre/gowk"
)

type GormBase struct {
}

func (m *GormBase) AutoMigrate(ctx context.Context, a ...any) error {
	return gowk.DB(ctx).AutoMigrate(a...)
}
