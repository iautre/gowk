package gowk

import (
	"log"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
)

type ServerConfig struct {
	HttpEngine *gin.Engine
	GrpcServer *GrpcServer
}

func Run(config *ServerConfig) {
	// Initialize database connection once
	initPostgres()

	// Start HTTP server first (for health checks and monitoring)
	if config.HttpEngine != nil {
		httpServer := &HttpServer{
			Engine: config.HttpEngine,
		}
		httpServer.ServerRun()
		log.Printf(" [INFO] HTTP server started")
	}

	// Start gRPC server second
	if config.GrpcServer != nil {
		config.GrpcServer.ServerRun()
		log.Printf(" [INFO] gRPC server started")
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Graceful shutdown
	log.Println("Shutting down servers...")

	// Shutdown gRPC server first (it stops immediately)
	if config.GrpcServer != nil {
		config.GrpcServer.ServerStop()
	}

	// Shutdown HTTP server second (with timeout)
	if config.HttpEngine != nil {
		httpServer := &HttpServer{
			Engine: config.HttpEngine,
		}
		httpServer.ServerStop()
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
