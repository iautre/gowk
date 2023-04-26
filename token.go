package gowk

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
)

const CONTEXT_LOGIN_ID_KEY = "CONTEXT_LOGIN_ID_KEY"
const TOKEN_KEY = "TOKEN_KEY"

var _defaultTokenHandler TokenHandler = &defaultToken{
	Token: make(map[string]*Token),
}
var _defaultTokenName string = "atoken"

type Token struct {
	Value   string `json:"value"`
	Name    string `json:"name"`
	Timeout int64  `json:"timeout"`
	loginId any    `json:"-"`
}

func CheckLoginMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenValue := ctx.Request.Header.Get(_defaultTokenName)
		if tokenValue == "" {
			Response().Fail(ctx, ERR_TOKEN)
			return
		}
		t, err := _defaultTokenHandler.GetToken(tokenValue)
		if err != nil {
			Response().Fail(ctx, ERR_TOKEN)
			return
		}
		if t == nil {
			Response().Fail(ctx, ERR_TOKEN)
			return
		}
		ctx.Next()
	}
}

func CheckLogin(ctx *gin.Context) {
	tokenValue := ctx.Request.Header.Get(_defaultTokenName)
	if tokenValue == "" {
		Response().Fail(ctx, ERR_TOKEN)
		return
	}
	t, err := _defaultTokenHandler.GetToken(tokenValue)
	if err != nil {
		Response().Fail(ctx, ERR_TOKEN)
		return
	}
	if t == nil {
		Response().Fail(ctx, ERR_TOKEN)
		return
	}
	ctx.Set(CONTEXT_LOGIN_ID_KEY, t.loginId)
	ctx.Next()
}

func SetTokenHandler(handler TokenHandler) {
	_defaultTokenHandler = handler
}
func SetTokenName(name string) {
	_defaultTokenName = name
}

func Login(ctx *gin.Context, loginId any) *Token {
	token := &Token{
		Value:   UUID(),
		Name:    _defaultTokenName,
		loginId: loginId,
	}
	_defaultTokenHandler.SaveToken(token.Value, token)
	return token
}
func GetLoginId(ctx context.Context) any {
	return ctx.Value(CONTEXT_LOGIN_ID_KEY)
}

type TokenHandler interface {
	SaveToken(key string, token *Token) error
	GetToken(key string) (*Token, error)
}

type defaultToken struct {
	Token map[string]*Token
}

func (d *defaultToken) SaveToken(key string, token *Token) error {
	d.Token[key] = token
	return nil
}
func (d *defaultToken) GetToken(key string) (*Token, error) {
	v, ok := d.Token[key]
	if !ok {
		return nil, errors.New("no token")
	}
	return v, nil
}
