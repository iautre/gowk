package gowk

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
)

type ServerConfig struct {
	HttpEngine *gin.Engine
	GrpcServer *GrpcServer
}

func Run(config *ServerConfig) {
	// Initialize database connection once
	initPostgres()

	// Store server references for graceful shutdown
	var httpServer *HttpServer
	var grpcServer *GrpcServer

	// Start HTTP server first (for health checks and monitoring)
	if config.HttpEngine != nil {
		httpServer = &HttpServer{
			Engine: config.HttpEngine,
		}
		httpServer.ServerRun() // This initializes Handler and starts server
		log.Printf(" [INFO] HTTP server started")
	}

	// Start gRPC server second
	if config.GrpcServer != nil {
		grpcServer = config.GrpcServer
		grpcServer.ServerRun()
		log.Printf(" [INFO] gRPC server started")
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Graceful shutdown with timeout
	log.Println("Shutting down servers...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown gRPC server first (it stops immediately)
	if grpcServer != nil {
		grpcServer.ServerStop()
		log.Printf(" [INFO] gRPC server stopped")
	}

	// Shutdown HTTP server second (with timeout)
	if httpServer != nil {
		if err := httpServer.Handler.Shutdown(ctx); err != nil {
			log.Printf(" [ERROR] HTTP server shutdown error: %v", err)
		} else {
			log.Printf(" [INFO] HTTP server stopped")
		}
	}

	// Close database connection once
	closePostgres()

	log.Println("All servers stopped")
}

// RunHTTP starts only HTTP server
func RunHTTP(engine *gin.Engine) {
	Run(&ServerConfig{
		HttpEngine: engine,
	})
}

// RunGRPC starts only gRPC server
func RunGRPC(grpcServer *GrpcServer) {
	Run(&ServerConfig{
		GrpcServer: grpcServer,
	})
}

// RunBoth starts both HTTP and gRPC servers
func RunBoth(engine *gin.Engine, grpcServer *GrpcServer) {
	Run(&ServerConfig{
		HttpEngine: engine,
		GrpcServer: grpcServer,
	})
}
