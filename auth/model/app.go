package model

import (
	"github.com/iautre/gowk"
	"github.com/jackc/pgtype"
)

type App struct {
	gowk.Model
	Key        string `gorm:"type:varchar;unique_index"`
	Name       string `gorm:"type:varchar"`
	Url        string `gorm:"type:varchar"`
	Type       string `gorm:"type:varchar"` //应用类型
	AuthIgnore bool   `gorm:"type:bool"`    //是否忽略需要认证
	AuthKey    string `gorm:"type:varchar"`
	AuthSecret string `gorm:"type:varchar"`
}

func (App) TableName() string {
	return "app"
}

type AppData struct {
	Id     uint64       `json:"id" gorm:"primaryKey"`
	Module string       `json:"module" gorm:"type:varchar"`
	Data   pgtype.JSONB `json:"data" gorm:"type:jsonb,serializer:json"`
	AppId  uint64       `json:"app_id" gorm:"type:bigint"` //应用id
}

func (AppData) TableName() string {
	return "app_data"
}
