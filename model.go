package gowk

import (
	"gorm.io/gorm"
)

type Model struct {
	ID      uint           `gorm:"primarykey" json:"-"`
	Created Time           `json:"created"`
	Updated Time           `json:"updated"`
	Deleted gorm.DeletedAt `gorm:"index" json:"-"`
}
