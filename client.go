package gowk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
)

const ContextClientKey = "AKEY_CONTEXT_CLIENT_KEY"
const redisClientPrefix = "AKEY_CLIENT_"

var _defaultClientHandler ClientHandler
var _defaultClientKeyNames = []string{"X-API-Key", "akey"}

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
	keyValue := clientKeyValue(ctx)
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
func SetClientKeyName(name string) {
	if name == "" {
		return
	}
	_defaultClientKeyNames = []string{name}
}
func SetClientKeyNames(names ...string) {
	var filtered []string
	for _, name := range names {
		if name == "" {
			continue
		}
		filtered = append(filtered, name)
	}
	if len(filtered) > 0 {
		_defaultClientKeyNames = filtered
	}
}

func clientKeyValue(ctx *gin.Context) string {
	for _, name := range _defaultClientKeyNames {
		if value := ctx.Request.Header.Get(name); value != "" {
			return value
		}
	}
	return ""
}

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
	jsonData, err := json.Marshal(client)
	if err != nil {
		return fmt.Errorf("marshal client: %w", err)
	}
	rdb := Redis()
	if rdb == nil {
		return errors.New("redis is not ready")
	}
	return rdb.Set(ctx, redisClientPrefix+key, string(jsonData), 0).Err()
}

func (d *redisClientStore) LoadClient(ctx context.Context, key string) (*Client, error) {
	rdb := Redis()
	if rdb == nil {
		return nil, errors.New("redis is not ready")
	}
	jsonData, err := rdb.Get(ctx, redisClientPrefix+key).Result()
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
