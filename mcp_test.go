package gowk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type testMCPProvider struct{}

func (testMCPProvider) MCPName() string                     { return "test" }
func (testMCPProvider) MCPVersion() string                  { return "v0.0.1" }
func (testMCPProvider) RegisterMCPTools(server *mcp.Server) {}

func TestMCPHandlerUsesStatelessTransport(t *testing.T) {
	handler := MCPHandler(testMCPProvider{})

	getReq := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET status = %d, want %d", getRec.Code, http.StatusMethodNotAllowed)
	}
	if allow := getRec.Header().Get("Allow"); allow != http.MethodPost {
		t.Fatalf("Allow = %q, want %q", allow, http.MethodPost)
	}
}

// localhostHostStatus 用真实的 HTTP 服务（监听 127.0.0.1）发一个 GET，
// 但把 Host 头伪造成外部域名，用来触发 SDK 的 DNS rebinding 保护分支。
func localhostHostStatus(t *testing.T, handler http.Handler) int {
	t.Helper()
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Host = "api.example.com"
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func TestMCPHandlerRejectsHostMismatchOnLoopback(t *testing.T) {
	if got := localhostHostStatus(t, MCPHandler(testMCPProvider{})); got != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (DNS rebinding protection)", got, http.StatusForbidden)
	}
}

func TestMCPHandlerWithOptionsDisablesLocalhostProtection(t *testing.T) {
	handler := MCPHandlerWithOptions(testMCPProvider{}, &mcp.StreamableHTTPOptions{
		Stateless:                  true,
		DisableLocalhostProtection: true,
	})
	// 关掉保护后不该再 403；无状态模式下 GET 仍是 405。
	if got := localhostHostStatus(t, handler); got != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", got, http.StatusMethodNotAllowed)
	}
}

type cancelMCPProvider struct {
	registrations atomic.Int32
	started       chan struct{}
	canceled      chan struct{}
}

func (*cancelMCPProvider) MCPName() string    { return "cancel-test" }
func (*cancelMCPProvider) MCPVersion() string { return "v0.0.1" }
func (p *cancelMCPProvider) RegisterMCPTools(server *mcp.Server) {
	p.registrations.Add(1)
	mcp.AddTool(server, &mcp.Tool{Name: "wait"}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		close(p.started)
		<-ctx.Done()
		close(p.canceled)
		return nil, nil, ctx.Err()
	})
}

func TestMCPHandlerSupportsLatestProtocolAndPropagatesCancellation(t *testing.T) {
	provider := &cancelMCPProvider{
		started:  make(chan struct{}),
		canceled: make(chan struct{}),
	}
	httpServer := httptest.NewTestServer(t, MCPHandler(provider))

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	session, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{
		Endpoint:             httpServer.URL,
		HTTPClient:           httpServer.Client(),
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("connect MCP client: %v", err)
	}
	defer session.Close()
	if got := session.InitializeResult().ProtocolVersion; got != "2026-07-28" {
		t.Fatalf("protocol version = %q, want %q", got, "2026-07-28")
	}

	callCtx, cancel := context.WithCancel(t.Context())
	callDone := make(chan error, 1)
	go func() {
		_, err := session.CallTool(callCtx, &mcp.CallToolParams{Name: "wait"})
		callDone <- err
	}()

	select {
	case <-provider.started:
	case <-time.After(5 * time.Second):
		t.Fatal("tool did not start")
	}
	cancel()

	select {
	case <-provider.canceled:
	case <-time.After(5 * time.Second):
		t.Fatal("HTTP cancellation did not reach tool context")
	}
	select {
	case <-callDone:
	case <-time.After(5 * time.Second):
		t.Fatal("canceled tool call did not return")
	}
	if got := provider.registrations.Load(); got != 1 {
		t.Fatalf("tool registrations = %d, want 1", got)
	}
}
