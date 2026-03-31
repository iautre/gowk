package gowk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
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
	jsonData, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("marshal token: %w", err)
	}
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
}
