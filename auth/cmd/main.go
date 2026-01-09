package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk"
	"github.com/iautre/gowk/auth"
	authpb "github.com/iautre/gowk/auth/proto"
)

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Parse command line flags with environment variable defaults
	httpPort := flag.String("http-port", getEnvOrDefault("HTTP_SERVER_ADDR", ":8087"), "HTTP server port")
	grpcPort := flag.String("grpc-port", getEnvOrDefault("GRPC_SERVER_ADDR", ":50051"), "gRPC server port")
	flag.Parse()

	// Set server addresses in gowk config
	gowk.SetHTTPServerAddr(*httpPort)
	gowk.SetGRPCServerAddr(*grpcPort)

	fmt.Printf("Starting servers with HTTP port: %s, gRPC port: %s\n", *httpPort, *grpcPort)

	// Create servers
	r := gowk.New()
	auth.Router(r.Group(gowk.AuthAPIPrefix()))

	grpcServer := gowk.NewGrpcServer()
	authServer := auth.NewAuthServer()
	authpb.RegisterAuthServiceServer(grpcServer.Server, authServer)

	// Add health check endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
			"services": gin.H{
				"http": "running",
				"grpc": "running",
			},
		})
	})

	r.GET("/grpc-status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"port":   "50051",
		})
	})

	// Start both servers using unified API
	gowk.RunBoth(r, grpcServer)
}
