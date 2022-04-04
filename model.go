package gowk

import (
	"gorm.io/gorm"
)

type Model struct {
	ID      uint           `gorm:"primarykey" json:"id"`
	Created Time           `json:"created"`
	Updated Time           `json:"updated"`
	Deleted gorm.DeletedAt `gorm:"index" json:"-"`
}

type PageModel struct {
	Size  int         `json:"size" form:"size"`
	Page  int         `json:"page" form:"page"`
	Total int64       `json:"total"`
	List  interface{} `json:"list"`
}

type M = map[string]interface{}
type A = []interface{}
