package model

import (
	"github.com/iautre/gowk"
	"github.com/iautre/gowk/auth/constant"
)

type User struct {
	gowk.Model
	Phone    string `json:"phone" gorm:"type:varchar"`
	Email    string `json:"email" gorm:"type:varchar"`
	Nickname string `json:"nickname" gorm:"type:varchar"`
	Group    string `json:"group" gorm:"type:varchar"`
	Status   uint   `json:"status" gorm:"type:tinyint;default:1"`
	Secret   string `json:"-" gorm:"type:varchar"` //2FA secret
}

func (u *User) TableName() string {
	return "user"
}

func NewUser(phone, email, nickname string) *User {
	return &User{
		Phone:    phone,
		Email:    email,
		Nickname: nickname,
		Group:    constant.USER_GROUP_DEFAULT,
		Status:   constant.ENABLE,
	}
}

type TokenInfo struct {
	Token string
}
