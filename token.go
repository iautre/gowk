package gowk

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
)

const CONTEXT_TOKEN_KEY = "ATOKEN_CONTEXT_TOKEN_KEY"
const CONTEXT_TOKEN_VALUE_KEY = "ATOKEN_CONTEXT_TOKEN_VALUE_KEY"
const CONTEXT_LOGIN_ID_KEY = "ATOKEN_CONTEXT_LOGIN_ID_KEY"

var _defaultTokenHandler TokenHandler = &defaultToken{
	Token: make(map[string]*Token),
}
var _defaultTokenName string = "atoken"
var _defaultTokenTimeout int64 = 30 * 24 * 60 * 60 //默认为秒/-1为永久有效

type Token struct {
	Value   string `json:"value"`
	Name    string `json:"name"`
	Timeout int64  `json:"timeout"`
	LoginId any    `json:"loginId"`
	Device  string `json:"device"`
}

func CheckLoginMiddleware() gin.HandlerFunc {
	return CheckLogin
}

func CheckLogin(ctx *gin.Context) {
	tokenValue := ctx.Request.Header.Get(_defaultTokenName)
	if tokenValue == "" {
		Panic(ERR_TOKEN)
	}
	token, err := _defaultTokenHandler.LoadToken(tokenValue)
	if err != nil {
		Panic(ERR_TOKEN)
	}
	if token == nil {
		Panic(ERR_TOKEN)
	}
	token.setContextToken(ctx, token)
	ctx.Next()
}

func SetTokenHandler(handler TokenHandler) {
	_defaultTokenHandler = handler
}
func SetTokenName(name string) {
	_defaultTokenName = name
}

// 默认为秒
func SetTokenTimeout(timeout int64) {
	_defaultTokenTimeout = timeout
}

type longIdType interface {
	string | int | uint | int64 | uint64
}

func Login[T longIdType](ctx *gin.Context, loginId T) string {
	token := &Token{
		Value:   UUID(),
		Name:    _defaultTokenName,
		Timeout: _defaultTokenTimeout,
		LoginId: loginId,
	}
	token.setContextToken(ctx, token)
	_defaultTokenHandler.StoreToken(token.Value, token)
	return token.Value
}
func (t *Token) setContextToken(ctx *gin.Context, token *Token) {
	ctx.Set(CONTEXT_TOKEN_KEY, token)
	ctx.Set(CONTEXT_TOKEN_VALUE_KEY, token.Value)
	ctx.Set(CONTEXT_LOGIN_ID_KEY, token.LoginId)
}
func TokenValue(ctx context.Context) string {
	return ctx.Value(CONTEXT_TOKEN_VALUE_KEY).(string)
}
func TokenInfo(ctx context.Context) *Token {
	return ctx.Value(CONTEXT_TOKEN_VALUE_KEY).(*Token)
}
func LoginId[T longIdType](ctx context.Context) T {
	var tmp T
	switch ctx.Value(CONTEXT_LOGIN_ID_KEY).(type) {
	case T:
		return ctx.Value(CONTEXT_LOGIN_ID_KEY).(T)
	default:
		return tmp
	}
}

type TokenHandler interface {
	StoreToken(tokenValue string, token *Token) error
	LoadToken(tokenValue string) (*Token, error)
}

type defaultToken struct {
	Token map[string]*Token
}

func (d *defaultToken) StoreToken(key string, token *Token) error {
	d.Token[key] = token
	return nil
}
func (d *defaultToken) LoadToken(key string) (*Token, error) {
	v, ok := d.Token[key]
	if !ok {
		return nil, errors.New("no token")
	}
	return v, nil
}
