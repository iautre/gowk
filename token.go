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
	return CheckLogin
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

type longIdType interface {
	string | int | uint | int64 | uint64
}

func Login[T longIdType](ctx *gin.Context, loginId T) *Token {
	token := &Token{
		Value:   UUID(),
		Name:    _defaultTokenName,
		loginId: loginId,
	}
	_defaultTokenHandler.SaveToken(token.Value, token)
	return token
}

func GetLoginId[T longIdType](ctx context.Context) T {
	var tmp T
	switch ctx.Value(CONTEXT_LOGIN_ID_KEY).(type) {
	case T:
		return ctx.Value(CONTEXT_LOGIN_ID_KEY).(T)
	default:
		return tmp
	}
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
