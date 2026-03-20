package gowk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

const ContextTokenKey = "ATOKEN_CONTEXT_TOKEN_KEY"
const ContextTokenValueKey = "ATOKEN_CONTEXT_TOKEN_VALUE_KEY"
const ContextLoginIdKey = "ATOKEN_CONTEXT_LOGIN_ID_KEY"
const ContextBasicAuthKey = "ATOKEN_BASIC_AUTH_KEY"

var _defaultTokenHandler TokenHandler
var _defaultTokenTimeout int64 = 30 * 24 * 60 * 60

// _basicAuthValidator 由业务层注册，用于校验 Basic Auth 凭据。
// 若未注册则拒绝所有 Basic Auth 请求。
var _basicAuthValidator func(decoded string) bool

// SetBasicAuthValidator 注册 Basic Auth 凭据校验函数。
func SetBasicAuthValidator(f func(decoded string) bool) {
	_basicAuthValidator = f
}

type Token struct {
	Value     string `json:"value"`
	Name      string `json:"name"`
	Timeout   int64  `json:"timeout"`
	LoginId   int64  `json:"loginId"`
	Device    string `json:"device"`
	CreatedAt int64  `json:"createdAt"`
}

func CheckLoginMiddleware() gin.HandlerFunc {
	return CheckLogin
}

func CheckLogin(ctx *gin.Context) {
	var token *Token
	var err error

	authHeader := ctx.GetHeader("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenValue := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenValue != "" {
				token, err = _defaultTokenHandler.LoadToken(ctx, tokenValue)
				if err == nil && token != nil {
					token.setContextToken(ctx, token)
					ctx.Next()
					return
				}
			}
		}
		if strings.HasPrefix(authHeader, "Basic ") {
			auth := strings.TrimPrefix(authHeader, "Basic ")
			if auth != "" {
				decoded, decErr := base64.StdEncoding.DecodeString(auth)
				if decErr == nil && _basicAuthValidator != nil && _basicAuthValidator(string(decoded)) {
					ctx.Set(ContextBasicAuthKey, string(decoded))
					ctx.Next()
					return
				}
			}
		}
	}

	oidcJwt, cookieErr := ctx.Cookie("oidc_jwt")
	if cookieErr == nil && oidcJwt != "" {
		token, err = _defaultTokenHandler.LoadToken(ctx, oidcJwt)
		if err == nil && token != nil {
			token.setContextToken(ctx, token)
			ctx.Next()
			return
		}
	}

	Response(ctx, http.StatusUnauthorized, nil, NewError("Authentication required"))
}

func SetTokenHandler(handler TokenHandler) {
	_defaultTokenHandler = handler
}


func SetTokenTimeout(timeout int64) {
	_defaultTokenTimeout = timeout
}

func Login(ctx *gin.Context, loginId int64) (string, error) {
	token := &Token{
		Value:     UUID(),
		Name:      "Bearer",
		Timeout:   _defaultTokenTimeout,
		LoginId:   loginId,
		CreatedAt: time.Now().Unix(),
	}
	loginWithOidcJwt(ctx, token)
	token.setContextToken(ctx, token)
	err := _defaultTokenHandler.StoreToken(ctx, token.Value, token)
	if err != nil {
		return "", err
	}
	return token.Value, nil
}

func loginWithOidcJwt(ctx *gin.Context, token *Token) {
	ctx.SetCookie("oidc_jwt", token.Value, 86400, "/", "", true, true)
}

func (t *Token) setContextToken(ctx *gin.Context, token *Token) {
	ctx.Set(ContextTokenKey, token)
	ctx.Set(ContextTokenValueKey, token.Value)
	ctx.Set(ContextLoginIdKey, token.LoginId)
}

// TokenValue 安全获取 token 字符串值，不存在时返回空字符串。
func TokenValue(ctx context.Context) string {
	v, _ := ctx.Value(ContextTokenValueKey).(string)
	return v
}

func TokenInfo(ctx context.Context) *Token {
	t, _ := ctx.Value(ContextTokenKey).(*Token)
	return t
}

// LoginId 安全获取登录 ID，不存在时返回 0。
func LoginId(ctx context.Context) int64 {
	v, _ := ctx.Value(ContextLoginIdKey).(int64)
	return v
}

type TokenHandler interface {
	StoreToken(context.Context, string, *Token) error
	LoadToken(context.Context, string) (*Token, error)
}

// defaultTokenStore 使用读写锁保护并发访问，并在 LoadToken 时检查过期。
type defaultTokenStore struct {
	mu    sync.RWMutex
	Token map[string]*Token
}

func (d *defaultTokenStore) StoreToken(_ context.Context, key string, token *Token) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Token[key] = token
	return nil
}

func (d *defaultTokenStore) LoadToken(_ context.Context, key string) (*Token, error) {
	d.mu.RLock()
	v, ok := d.Token[key]
	d.mu.RUnlock()
	if !ok {
		return nil, errors.New("no token")
	}
	if v.Timeout > 0 && v.CreatedAt > 0 {
		if time.Now().Unix() > v.CreatedAt+v.Timeout {
			d.mu.Lock()
			delete(d.Token, key)
			d.mu.Unlock()
			return nil, errors.New("token expired")
		}
	}
	return v, nil
}

type redisTokenStore struct{}

const redisTokenPrefix = "ATOKEN_TOKEN_"

func (d *redisTokenStore) StoreToken(ctx context.Context, key string, token *Token) error {
	jsonData, _ := json.Marshal(token)
	return Redis().Set(ctx, redisTokenPrefix+key, string(jsonData), time.Duration(_defaultTokenTimeout)*time.Second).Err()
}

func (d *redisTokenStore) LoadToken(ctx context.Context, key string) (*Token, error) {
	jsonData, err := Redis().Get(ctx, redisTokenPrefix+key).Result()
	if err != nil {
		return nil, err
	}
	var token Token
	if err := json.Unmarshal([]byte(jsonData), &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func init() {
	if HasRedis() {
		_defaultTokenHandler = &redisTokenStore{}
	} else {
		_defaultTokenHandler = &defaultTokenStore{
			Token: make(map[string]*Token),
		}
	}
	initWeapp()
}

// ── 微信相关 ──────────────────────────────────────────────────────────────────

func initWeapp() {
	if HasWeapp() {
		go func() {
			var weapp Weapp
			weapp.InitWeapp()
			ticker := time.NewTicker((7200 - 60) * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				weapp.InitWeapp()
			}
		}()
	}
}

func GetWeappAccessToken() string {
	if weapp_access_token.Load() == nil {
		return ""
	}
	return weapp_access_token.Load().AccessToken
}

func GetWeappJsapiTicket() string {
	if weapp_jsapi_ticket.Load() == nil {
		return ""
	}
	return weapp_jsapi_ticket.Load().Ticket
}

var (
	weapp_access_token atomic.Pointer[WeappAccessToken]
	weapp_jsapi_ticket atomic.Pointer[WeappJsapiTicket]
)

type Weapp struct {
	AccessToken string `json:"access_token"`
	Ticket      string `json:"ticket"`
}

type WeappErr struct {
	Errcode int64  `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

type WeappAccessToken struct {
	WeappErr
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ExpiresTime int64
}

type WeappJsapiTicket struct {
	WeappErr
	Ticket      string `json:"ticket"`
	ExpiresIn   int64  `json:"expires_in"`
	ExpiresTime int64
}

func (w *Weapp) InitWeapp() {
	slog.Info("获取微信access_token")
	if err := w.SetAccessToken(); err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info("获取微信jsapi_ticket")
	if err := w.SetJsapiTicket(); err != nil {
		slog.Error(err.Error())
	}
}

func (w *Weapp) SetAccessToken() error {
	wt, err := w.GetAccessToken(context.TODO())
	if err != nil {
		return err
	}
	weapp_access_token.Store(wt)
	return nil
}

func (w *Weapp) SetJsapiTicket() error {
	if !weappJsapiTicket {
		return nil
	}
	if weapp_access_token.Load() == nil {
		return errors.New("jsapi_ticket初始化失败：access_token未就绪")
	}
	wt, err := w.GetJsapiTicket(context.TODO(), weapp_access_token.Load().AccessToken)
	if err != nil {
		return err
	}
	weapp_jsapi_ticket.Store(wt)
	return nil
}

const getAccessTokenUrl = "https://api.weixin.qq.com/cgi-bin/token"

func (w *Weapp) GetAccessToken(ctx context.Context) (*WeappAccessToken, error) {
	if !HasWeapp() {
		return nil, errors.New("weapp配置错误")
	}
	res, err := HttpClient().Get(fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s",
		getAccessTokenUrl, weappAppid, weappSecret))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var t WeappAccessToken
	if err = json.NewDecoder(res.Body).Decode(&t); err != nil {
		return nil, err
	}
	if t.Errcode != 0 {
		return nil, fmt.Errorf("errcode:%d, errmsg: %s", t.Errcode, t.Errmsg)
	}
	t.ExpiresTime = t.ExpiresIn + time.Now().Unix()
	return &t, nil
}

const getJsapiTicketUrl = "https://api.weixin.qq.com/cgi-bin/ticket/getticket"

func (w *Weapp) GetJsapiTicket(ctx context.Context, accessToken string) (*WeappJsapiTicket, error) {
	res, err := HttpClient().Get(fmt.Sprintf("%s?access_token=%s&type=jsapi", getJsapiTicketUrl, accessToken))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var t WeappJsapiTicket
	if err = json.NewDecoder(res.Body).Decode(&t); err != nil {
		return nil, err
	}
	if t.Errcode != 0 {
		return nil, fmt.Errorf("errcode:%d, errmsg: %s", t.Errcode, t.Errmsg)
	}
	t.ExpiresTime = t.ExpiresIn + time.Now().Unix()
	return &t, nil
}
