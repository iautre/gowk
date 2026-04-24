package gowk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RouteInfo struct {
	Method      string
	Path        string
	Handler     string
	HandlerFunc func(context.Context)
}

type HttpServer struct {
	Handler *http.Server
	Engine  *gin.Engine
}

func New() *gin.Engine {
	slog.SetDefault(Logger(slog.LevelInfo))
	engine := gin.New()
	engine.Use(GlobalErrorHandler(), LogTrace(), Recover(), TransactionHandler())
	engine.NoRoute(NotFound())
	engine.NoMethod(NotFound())
	return engine
}

// ServerRun 同步执行 net.Listen 绑定端口，成功后把 Serve 放到 goroutine 里运行。
// 监听失败（端口占用/地址非法等）直接返回 error，由调用方决定 fail-fast；
// Serve 阶段的非 ErrServerClosed 错误只打日志，不再向上传递（此时端口已绑好，进程也已通告 running）。
func (h *HttpServer) ServerRun() error {
	if h.Engine == nil {
		h.Engine = gin.Default()
	}
	if h.Handler == nil {
		h.Handler = &http.Server{
			Addr:    httpServerAddr,
			Handler: h.Engine,
		}
	}
	ln, err := net.Listen("tcp", h.Handler.Addr)
	if err != nil {
		return fmt.Errorf("HTTP 监听失败 addr=%s: %w", h.Handler.Addr, err)
	}
	slog.Info("HTTP server running", "addr", ln.Addr().String())
	go func() {
		if err := h.Handler.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server serve failed", "addr", ln.Addr().String(), "err", err)
		}
	}()
	return nil
}

func (h *HttpServer) ServerStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.Handler.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		// 不走 Fatal，避免跳过 defer 导致资源无法释放
		slog.Error("HTTP server shutdown failed", "err", err)
	}
	slog.Info("HTTP server stopped")
}
