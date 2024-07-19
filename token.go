package gowk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iautre/gowk/conf"
	"net/http"
	"time"

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

/**
*微信相关
 */

type Weapp struct {
	appid  string
	secret string
}

type WeappErr struct {
	Errcode int64  `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

type WeappToken struct {
	WeappErr
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ExpiresTime int64
}

const getAccessTokenUrl = "https://api.weixin.qq.com/cgi-bin/token"

func NewWeapp() *Weapp {
	temp := conf.Get[map[string]any]("weapp")
	return &Weapp{
		appid:  (*temp)["appid"].(string),
		secret: (*temp)["secret"].(string),
	}
}

func (w *Weapp) GetAccessToken(ctx context.Context) (*WeappToken, error) {
	res, err := http.Get(fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s", getAccessTokenUrl, w.appid, w.secret))
	if err != nil {
		return nil, err
	}
	var t WeappToken
	err = json.NewDecoder(res.Body).Decode(&t)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	if t.Errcode != 0 {
		return nil, err
	}
	t.ExpiresTime = t.ExpiresIn + time.Now().Unix()
	return &t, nil
}

type WeappJsapiTicket struct {
	WeappErr
	Ticket      string `json:"ticket"`
	ExpiresIn   int64  `json:"expires_in"`
	ExpiresTime int64
}

const getJsapiTicketUrl = "https://api.weixin.qq.com/cgi-bin/ticket/getticket"

func (w *Weapp) GetJsapiTicket(ctx context.Context, accessToken string) (*WeappJsapiTicket, error) {
	res, err := http.Get(fmt.Sprintf("%s?access_token=%s&type=jsapi", getJsapiTicketUrl, accessToken))
	if err != nil {
		return nil, err
	}
	var t WeappJsapiTicket
	err = json.NewDecoder(res.Body).Decode(&t)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	if t.Errcode != 0 {
		return nil, err
	}
	t.ExpiresTime = t.ExpiresIn + time.Now().Unix()
	return &t, nil
}
