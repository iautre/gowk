package gowk

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

type ServerConfig struct {
	HttpEngine *gin.Engine
	GrpcServer *GrpcServer
}

func Run(config *ServerConfig) {
	// 触发 Postgres 后台初始化（非阻塞，连不上也不退出，后台退避重试）。
	// Redis 保持按需：首次 Redis() / InitRedis() 时才触发后台初始化。
	InitPostgres()

	var httpServer *HttpServer
	var grpcServer *GrpcServer

	if config.HttpEngine != nil {
		httpServer = &HttpServer{Engine: config.HttpEngine}
		if err := httpServer.ServerRun(); err != nil {
			slog.Error("HTTP 启动失败，进程退出", "err", err)
			// 监听都没成功，无需 Shutdown HTTP；顺手清理已触发的依赖初始化。
			closePostgres()
			closeRedis()
			os.Exit(1)
		}
	}

	if config.GrpcServer != nil {
		grpcServer = config.GrpcServer
		if err := grpcServer.ServerRun(); err != nil {
			slog.Error("gRPC 启动失败，进程退出", "err", err)
			// HTTP 可能已经起来了，先优雅关掉避免端口残留。
			if httpServer != nil {
				httpServer.ServerStop()
			}
			closePostgres()
			closeRedis()
			os.Exit(1)
		}
	}

	quit := make(chan os.Signal, 1)
	// 同时监听 SIGINT（Ctrl+C）和 SIGTERM（Docker/K8s 停止信号）
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down servers...")

	if grpcServer != nil {
		grpcServer.ServerStop()
		slog.Info("gRPC server stopped")
	}
	if httpServer != nil {
		httpServer.ServerStop()
		slog.Info("HTTP server stopped")
	}

	closePostgres()
	closeRedis()
	slog.Info("All servers stopped")
}

func RunHTTP(engine *gin.Engine) {
	Run(&ServerConfig{HttpEngine: engine})
}

func RunGRPC(grpcServer *GrpcServer) {
	Run(&ServerConfig{GrpcServer: grpcServer})
}

func RunBoth(engine *gin.Engine, grpcServer *GrpcServer) {
	Run(&ServerConfig{HttpEngine: engine, GrpcServer: grpcServer})
}

// InitPostgres 触发 Postgres 的 sync.Once 初始化路径。
// 只会启动后台重试 goroutine，本身非阻塞；连接失败不会退出进程，
// 在连接成功之前 Postgres(ctx) 返回 nil，PostgresTx 返回 "postgres unavailable"。
func InitPostgres() {
	Postgres(context.Background())
}

// InitRedis 触发 Redis 的 sync.Once 初始化路径。
// 只会启动后台重试 goroutine，本身非阻塞；连接失败不会退出进程，
// 在连接成功之前 Redis() 返回 nil。
func InitRedis() {
	Redis()
}
