package gowk

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const ContextClientKey = "AKEY_CONTEXT_CLIENT_KEY"
const redisClientPrefix = "AKEY_CLIENT_"

var _defaultClientHandler ClientHandler
var _defaultClientKeyName = "akey"

type Client struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	LoginId int64  `json:"loginId"`
	Device  string `json:"device"`
}

func CheckClientMiddleware() gin.HandlerFunc {
	return CheckClient
}

func CheckClient(ctx *gin.Context) {
	keyValue := ctx.Request.Header.Get(_defaultClientKeyName)
	if keyValue == "" {
		ctx.Error(ERR_AUTH)
		ctx.Abort()
		return
	}
	client, err := _defaultClientHandler.LoadClient(ctx, keyValue)
	if err != nil || client == nil {
		ctx.Error(ERR_AUTH)
		ctx.Abort()
		return
	}
	client.setContextClient(ctx, client)
	ctx.Next()
}

func SetClientHandler(handler ClientHandler) { _defaultClientHandler = handler }
func SetClientKeyName(name string)           { _defaultClientKeyName = name }

func (t *Client) setContextClient(ctx *gin.Context, client *Client) {
	ctx.Set(ContextClientKey, client)
	ctx.Set(ContextLoginIdKey, client.LoginId)
}

type ClientHandler interface {
	StoreClient(context.Context, string, *Client) error
	LoadClient(context.Context, string) (*Client, error)
}

func StoreClient(ctx context.Context, key string, client *Client) error {
	return _defaultClientHandler.StoreClient(ctx, key, client)
}

// defaultClientStore 使用读写锁保护并发访问。
type defaultClientStore struct {
	mu     sync.RWMutex
	Client map[string]*Client
}

func (d *defaultClientStore) StoreClient(_ context.Context, key string, client *Client) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Client[key] = client
	return nil
}

func (d *defaultClientStore) LoadClient(_ context.Context, key string) (*Client, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	v, ok := d.Client[key]
	if !ok {
		return nil, errors.New("no client")
	}
	return v, nil
}

type redisClientStore struct{}

func (d *redisClientStore) StoreClient(ctx context.Context, key string, client *Client) error {
	jsonData, _ := json.Marshal(client)
	return Redis().Set(ctx, redisClientPrefix+key, string(jsonData), time.Duration(_defaultTokenTimeout)*time.Second).Err()
}

func (d *redisClientStore) LoadClient(ctx context.Context, key string) (*Client, error) {
	jsonData, err := Redis().Get(ctx, redisClientPrefix+key).Result()
	if err != nil {
		return nil, err
	}
	var client Client
	if err := json.Unmarshal([]byte(jsonData), &client); err != nil {
		return nil, err
	}
	return &client, nil
}

func init() {
	if HasRedis() {
		_defaultClientHandler = &redisClientStore{}
	} else {
		_defaultClientHandler = &defaultClientStore{
			Client: make(map[string]*Client),
		}
	}
}
