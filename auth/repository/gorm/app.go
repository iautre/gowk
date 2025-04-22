package gorm

import (
	"context"
	"github.com/iautre/gowk"
	"github.com/iautre/gowk/auth/model"
)

type GormApp struct{}

func (a *GormApp) GetByName(ctx context.Context, name string) (*model.App, error) {
	var app model.App
	if err := gowk.DB(ctx).WithContext(ctx).Where("name = ?", name).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (a *GormApp) GetByKey(ctx context.Context, key string) (*model.App, error) {
	var app model.App
	if err := gowk.DB(ctx).WithContext(ctx).Where("key = ?", key).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

type GormAppData struct{}

func (d *GormAppData) GetById(ctx context.Context, appId uint64, module string, id uint64) (*model.AppData, error) {
	var res model.AppData
	if err := gowk.DB(ctx).WithContext(ctx).Where("app_id = ?", appId).Where("module = ?", module).Where("id = ?", id).First(&res).Error; err != nil {
		return nil, err
	}
	return &res, nil
}

func (d *GormAppData) Get(ctx context.Context, appId uint64, module string, params gowk.M) ([]*model.AppData, error) {
	var ads []*model.AppData
	db := gowk.DB(ctx).WithContext(ctx).Where("app_id = ?", appId).Where("module = ?", module)
	if len(params) > 0 {
		for k, v := range params {
			db = db.Where("data ->> ? = ?", k, v)
		}
	}
	if result := db.Find(&ads); result.Error != nil {
		return nil, result.Error
	}
	return ads, nil
}

func (d *GormAppData) Save(ctx context.Context, appId uint64, table string, data gowk.M) (uint64, error) {
	return 0, nil
}
func (d *GormAppData) Delete(ctx context.Context, appId uint64, table string, id uint64) error {
	return nil
}
