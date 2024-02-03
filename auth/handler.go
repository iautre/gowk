package auth

import (
	"context"

	"github.com/iautre/gowk"
)

type defaultToken struct {
}

func (d *defaultToken) StoreToken(key string, token *gowk.Token) error {
	var userService UserService
	userService.UpdateToken(context.TODO(), token.LoginId.(int64), key)
	return nil
}
func (d *defaultToken) LoadToken(key string) (*gowk.Token, error) {
	var userService UserService
	user := userService.GetByToken(context.TODO(), key)
	return &gowk.Token{
		Value:   key,
		LoginId: user.Id,
	}, nil
}
func init() {
	gowk.SetTokenHandler(&defaultToken{})
}
