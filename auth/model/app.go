package model

import (
	"github.com/jackc/pgtype"
)

type App struct {
	Id         uint64 `json:"id" gorm:"primaryKey"`
	Key        string
	Name       string
	Url        string
	Type       string //应用类型
	AuthIgnore bool   //是否忽略需要认证
	AuthKey    string
	AuthSecret string
}

func (App) TableName() string {
	return "app"
}

type AppData struct {
	Id     uint64       `json:"id" gorm:"primaryKey"`
	Module string       `json:"module"`
	Data   pgtype.JSONB `json:"data" gorm:"type:jsonb,serializer:json"`
	AppId  uint64       //应用id
}

func (AppData) TableName() string {
	return "app_data"
}
