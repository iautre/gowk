package gowk

import (
	"context"
	"log"
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
	// 通过 sync.Once 保证连接池只初始化一次
	InitPostgres()

	var httpServer *HttpServer
	var grpcServer *GrpcServer

	if config.HttpEngine != nil {
		httpServer = &HttpServer{Engine: config.HttpEngine}
		httpServer.ServerRun()
		log.Printf(" [INFO] HTTP server started")
	}

	if config.GrpcServer != nil {
		grpcServer = config.GrpcServer
		grpcServer.ServerRun()
		log.Printf(" [INFO] gRPC server started")
	}

	quit := make(chan os.Signal, 1)
	// 同时监听 SIGINT（Ctrl+C）和 SIGTERM（Docker/K8s 停止信号）
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	if grpcServer != nil {
		grpcServer.ServerStop()
		log.Printf(" [INFO] gRPC server stopped")
	}
	if httpServer != nil {
		httpServer.ServerStop()
		log.Printf(" [INFO] HTTP server stopped")
	}

	closePostgres()
	log.Println("All servers stopped")
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

// InitPostgres 供外部或 Run() 调用，通过 sync.Once 保证只初始化一次。
func InitPostgres() {
	// 触发 Postgres() 内的 pgInitOnce.Do(initPostgres)
	Postgres(context.Background())
}
