package auth

import (
	"github.com/iautre/gowk"
)

type User struct {
	Id       int64      `json:"id"`
	Phone    string     `json:"phone"`
	Email    string     `json:"email"`
	Nickname string     `json:"nickname"`
	Group    string     `json:"group"`
	Status   uint       `json:"status"`
	Created  *gowk.Time `json:"created"`
	Updated  *gowk.Time `json:"updated"`
	Secret   string     `json:"-"` //2FA secret
}

func (u *User) TableName() string {
	return "user"
}

func NewUser(phone, email, nickname string) *User {
	return &User{
		Phone:    phone,
		Email:    email,
		Nickname: nickname,
		Group:    USER_GROUP_DEFAULT,
		Status:   ENABLE,
		Created:  gowk.Now(),
		Updated:  gowk.Now(),
	}
}

type TokenInfo struct {
	Token string
}
