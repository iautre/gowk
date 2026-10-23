package gowk

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPProvider 由业务方实现：提供 MCP 服务标识与 tool 注册逻辑。
// tool 内部直接调用业务的 internal/service（与 HTTP handler 同一份逻辑），
// gowk 只负责把它以 Streamable HTTP 挂到现有 gin 服务并复用端口与鉴权。
type MCPProvider interface {
	MCPName() string
	MCPVersion() string
	RegisterMCPTools(server *mcp.Server)
}

// newMCPServer 由 provider 构建一个已注册 tool 的 MCP server。
func newMCPServer(p MCPProvider) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: p.MCPName(), Version: p.MCPVersion()}, nil)
	p.RegisterMCPTools(server)
	return server
}

// MCPHandler 返回 MCP SDK 的无状态 Streamable HTTP 处理器。
func MCPHandler(p MCPProvider) http.Handler {
	return MCPHandlerWithOptions(p, nil)
}

// MCPHandlerWithOptions 同 MCPHandler，但允许自定义 Streamable HTTP 选项；
// opts 为 nil 时使用默认值（Stateless + 请求取消传播）。
//
// 注意：MCP SDK 自 v1.8.0 起默认开启 DNS rebinding 保护——连接落在回环地址、
// 而 Host 头不是回环地址的请求会被直接 403。若服务挂在同机反向代理后面且代理
// 透传了真实域名（如 nginx 的 proxy_set_header Host $host），需要显式设置
// opts.DisableLocalhostProtection = true，或让代理把 Host 改写为 127.0.0.1。
func MCPHandlerWithOptions(p MCPProvider, opts *mcp.StreamableHTTPOptions) http.Handler {
	server := newMCPServer(p)
	if opts == nil {
		opts = &mcp.StreamableHTTPOptions{
			Stateless:                    true,
			PropagateRequestCancellation: true,
		}
	}
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, opts)
}

// SetupMCP 把 provider 的 MCP（Streamable HTTP）挂到现有 gin 路由 path，复用主服务端口。
// mw 为前置中间件：MCP 与其对应的 HTTP 接口套用同一道鉴权，校验不过会 Abort，
// 请求不会进入 MCP handler，因此 MCP tool 不必自行鉴权。
//
// 这里不用 gin.WrapH，而是自行 ServeHTTP：把登录中间件写入 gin.Context 的鉴权身份
// 透传进 request context，使 MCP tool（参数为 context.Context）可经 gowk.LoginId /
// TokenValue 读取调用方身份，从而直连需要身份的 service。
//
//	gowk.SetupMCP(router, "/mcp", provider, mw.Login)
func SetupMCP(r gin.IRouter, path string, p MCPProvider, mw ...gin.HandlerFunc) {
	SetupMCPWithOptions(r, path, p, nil, mw...)
}

// SetupMCPWithOptions 同 SetupMCP，但允许自定义 Streamable HTTP 选项。
// 用法见 MCPHandlerWithOptions 的说明（同机反向代理场景需要关掉 DNS rebinding 保护）。
func SetupMCPWithOptions(r gin.IRouter, path string, p MCPProvider, opts *mcp.StreamableHTTPOptions, mw ...gin.HandlerFunc) {
	h := MCPHandlerWithOptions(p, opts)
	handlers := make([]gin.HandlerFunc, 0, len(mw)+1)
	handlers = append(handlers, mw...)
	handlers = append(handlers, func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request.WithContext(mcpRequestContext(c)))
	})
	r.Any(path, handlers...)
}

// mcpRequestContext 把前置中间件写入 gin.Context 的所有键值（登录 ID、token、*User 等）
// 透传到 request context，使 MCP tool（参数为 context.Context）能像 HTTP handler 一样
// 通过 ctx.Value 读取鉴权身份，从而直连依赖身份的 service。不感知具体业务键。
func mcpRequestContext(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	for k, v := range c.Keys {
		ctx = context.WithValue(ctx, k, v)
	}
	return ctx
}
