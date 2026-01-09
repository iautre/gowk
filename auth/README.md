# Auth Module

这是一个完整的OAuth2/OIDC认证模块，既可以作为独立服务端运行，也可以作为客户端库被其他应用引用。

## 架构设计

### 双重用途
- ✅ **服务端模式**: 独立运行提供完整的OAuth2/OIDC服务
- ✅ **客户端模式**: 被其他应用import作为OAuth2客户端库
- ✅ **gRPC接口**: 内部高性能通信协议

### 模块结构
```
auth/
├── cmd/main.go              # 服务端入口
├── client.go               # OAuth2客户端库
├── service.go              # 业务逻辑层
├── handler.go              # HTTP处理器
├── proto/                  # gRPC协议定义
├── sql/                    # 数据库查询
├── db/                     # 数据库模型
└── examples/                # 使用示例
```

## 功能特性

### 服务端功能
- ✅ OAuth2授权码流程 (Authorization Code Flow)
- ✅ PKCE支持 (Proof Key for Code Exchange)
- ✅ 刷新令牌 (Refresh Token)
- ✅ 获取用户信息 (User Info)
- ✅ 撤销令牌 (Revoke Token)
- ✅ OIDC支持 (OpenID Connect)
- ✅ RSA签名ID令牌
- ✅ TTL配置管理
- ✅ 标准化JSON响应

### 客户端功能
- ✅ gRPC高性能通信
- ✅ 自动令牌刷新
- ✅ PKCE安全增强
- ✅ 错误处理和重试
- ✅ 简洁易用的API

## 快速开始

### 作为服务端运行

```bash
# 编译服务端
go build -o auth-server ./auth/cmd/main.go

# 运行服务端
./auth-server
```

### 作为客户端库使用

```go
import "github.com/iautre/gowk/auth"

// 创建OAuth2客户端
client, err := auth.NewOAuth2Client(
    "localhost:50051", // gRPC服务地址
    "test-client-id",  // 客户端ID
    "test-client-secret", // 客户端密钥
    "http://localhost:3000/callback", // 回调地址
)
if err != nil {
    log.Fatalf("Failed to create OAuth2 client: %v", err)
}
defer client.Close()

// 获取访问令牌
tokenResult, err := client.GetToken(ctx, code, codeVerifier)
if err != nil {
    log.Fatalf("Get token failed: %v", err)
}

// 获取用户信息
userInfo, err := client.GetUserInfo(ctx, tokenResult.AccessToken)
if err != nil {
    log.Fatalf("Get user info failed: %v", err)
}
```

## 安装依赖

```bash
# 主项目依赖管理
go mod tidy

# 生成protobuf代码
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    auth/proto/auth.proto
```

## 使用示例

### 完整OAuth2流程

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/iautre/gowk/auth"
)

func main() {
    // 创建OAuth2客户端
    client, err := auth.NewOAuth2Client(
        "localhost:50051", // gRPC服务地址
        "test-client-id",  // 客户端ID
        "test-client-secret", // 客户端密钥
        "http://localhost:3000/callback", // 回调地址
    )
    if err != nil {
        log.Fatalf("Failed to create OAuth2 client: %v", err)
    }
    defer client.Close()

    ctx := context.Background()

    // 1. 获取授权码
    authResult, err := client.Authorize(ctx, "openid profile email", "random-state")
    if err != nil {
        log.Fatalf("Authorize failed: %v", err)
    }

    // 2. 获取访问令牌
    tokenResult, err := client.GetToken(ctx, authResult.Code, authResult.CodeVerifier)
    if err != nil {
        log.Fatalf("Get token failed: %v", err)
    }

    // 3. 获取用户信息
    userInfo, err := client.GetUserInfo(ctx, tokenResult.AccessToken)
    if err != nil {
        log.Fatalf("Get user info failed: %v", err)
    }

    log.Printf("User: %s (%s)", userInfo.Name, userInfo.Email)
}
```

### 自动令牌管理

```go
type tokenManager struct {
    client       *auth.AuthClient
    accessToken  string
    refreshToken string
    expiresAt    time.Time
}

func newTokenManager(client *auth.AuthClient) *tokenManager {
    return &tokenManager{
        client: client,
    }
}

func (tm *tokenManager) ensureValidToken(ctx context.Context) error {
    // 检查令牌是否即将过期（提前5分钟刷新）
    if time.Now().Add(5 * time.Minute).Before(tm.expiresAt) {
        return nil // 令牌仍然有效
    }

    // 刷新令牌
    result, err := tm.client.RefreshToken(ctx, tm.refreshToken)
    if err != nil {
        return fmt.Errorf("failed to refresh token: %w", err)
    }

    // 更新令牌信息
    tm.accessToken = result.AccessToken
    tm.refreshToken = result.RefreshToken
    tm.expiresAt = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

    return nil
}

func (tm *tokenManager) getUserInfo(ctx context.Context) (*auth.UserInfo, error) {
    if err := tm.ensureValidToken(ctx); err != nil {
        return nil, err
    }
    return tm.client.GetUserInfo(ctx, tm.accessToken)
}
```

## API参考

### AuthClient

#### NewAuthClient
```go
func NewOAuth2Client(serverAddr, clientID, clientSecret, redirectURI string) (*AuthClient, error)
```

#### Authorize
```go
func (c *AuthClient) Authorize(ctx context.Context, scope, state string) (*AuthorizeResult, error)
```

#### GetToken
```go
func (c *AuthClient) GetToken(ctx context.Context, code, codeVerifier string) (*TokenResult, error)
```

#### RefreshToken
```go
func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error)
```

#### GetUserInfo
```go
func (c *AuthClient) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error)
```

#### RevokeToken
```go
func (c *AuthClient) RevokeToken(ctx context.Context, token, tokenTypeHint string) (*RevokeResult, error)
```

### 数据结构

#### AuthorizeResult
```go
type AuthorizeResult struct {
    Code            string
    State           string
    RedirectURI     string
    CodeVerifier    string
    Error           string
    ErrorDescription string
}
```

#### TokenResult
```go
type TokenResult struct {
    AccessToken     string
    TokenType       string
    ExpiresIn       int64
    RefreshToken    string
    Scope           string
    IDToken         string
    Error           string
    ErrorDescription string
}
```

#### UserInfo
```go
type UserInfo struct {
    Sub               string
    Name              string
    Email             string
    Phone             string
    Picture           string
    Nickname          string
    PreferredUsername string
    UpdatedAt         string
    Error             string
    ErrorDescription  string
}
```

## 错误处理

所有API调用都返回错误信息，可以通过检查`Error`字段来判断是否成功：

```go
result, err := client.GetToken(ctx, code, verifier)
if err != nil {
    return fmt.Errorf("network error: %w", err)
}

if result.Error != "" {
    return fmt.Errorf("API error: %s - %s", result.Error, result.ErrorDescription)
}
```

## 安全注意事项

1. **客户端密钥安全**: 确保客户端密钥安全存储，不要硬编码在代码中
2. **HTTPS通信**: 生产环境中使用TLS加密gRPC通信
3. **令牌存储**: 安全存储访问令牌和刷新令牌
4. **PKCE使用**: 始终使用PKCE增强安全性
5. **状态参数**: 使用随机状态参数防止CSRF攻击

## 配置管理

### 环境变量
```bash
# 服务端配置
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_NAME="auth_db"
export DB_USER="postgres"
export DB_PASSWORD="password"

# API前缀配置
export AUTH_API_PREFIX="/api/v1"  # 可选，默认为/api/v1

# 客户端配置
export AUTH_SERVER_ADDR="localhost:50051"
export CLIENT_ID="your-client-id"
export CLIENT_SECRET="your-client-secret"
export REDIRECT_URI="http://localhost:3000/callback"
```

### 数据库配置

OAuth2客户端和JWK密钥通过数据库直接管理，支持：
- ✅ TTL配置: 每个客户端可配置不同的令牌过期时间
- ✅ 直接管理: 通过数据库工具直接操作，无需API
- ✅ 安全隔离: 敏感配置不暴露给应用层

### 服务端启动
```bash
# 开发模式
go run ./auth/cmd/main.go

# 生产模式
go build -o auth-server ./auth/cmd/main.go
./auth-server --port=8080 --db-host=localhost
```

## 运行示例

### 启动服务端
```bash
# 开发环境
cd auth
go run ./cmd/main.go

# 生产环境
go build -o auth-server ./cmd/main.go
./auth-server
```

### 运行客户端示例
```bash
cd auth/examples
go run basic_example.go
```

### 测试API
```bash
# 健康检查
curl http://localhost:8080/health

# OAuth2授权 (使用默认前缀 /api/v1)
curl -X POST "http://localhost:8080/api/v1/oauth2/auth" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"test","redirect_uri":"http://localhost:3000/callback","response_type":"code","scope":"openid profile"}'

# 获取令牌 (使用默认前缀 /api/v1)
curl -X POST "http://localhost:8080/api/v1/oauth2/token" \
  -H "Content-Type: application/json" \
  -d '{"grant_type":"authorization_code","code":"xxx","redirect_uri":"http://localhost:3000/callback"}'

# 自定义API前缀
export AUTH_API_PREFIX="/auth"
# 重启服务后，端点变为：
# http://localhost:8080/auth/oauth2/auth
# http://localhost:8080/auth/oauth2/token
```

## 依赖项

### 核心依赖
- **Go 1.25.5+**: 编译和运行环境
- **gRPC v1.78.0+**: 高性能RPC通信
- **Protocol Buffers v1.36.11+**: 序列化协议
- **Gin v1.11.0+**: HTTP服务框架
- **PGX v5.8.0+**: PostgreSQL数据库驱动

### 开发依赖
- **sqlc**: SQL代码生成
- **protoc**: Protocol Buffers编译器
- **PostgreSQL**: 数据库服务

### 生产依赖
- **Redis**: 缓存和会话存储
- **PostgreSQL**: 主数据存储
- **TLS证书**: HTTPS通信安全

## 性能特性

### 高性能设计
- ✅ **gRPC通信**: 二进制协议，高性能
- ✅ **连接池**: 数据库连接复用
- ✅ **缓存策略**: Redis缓存热点数据
- ✅ **异步处理**: 非阻塞I/O操作

### 安全特性
- ✅ **PKCE**: 授权码交换增强
- ✅ **RSA签名**: ID令牌数字签名
- ✅ **TTL管理**: 灵活的令牌过期策略
- ✅ **状态参数**: CSRF攻击防护
- ✅ **HTTPS**: 传输层加密

## 许可证

MIT License
