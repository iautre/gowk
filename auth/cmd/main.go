package main

import (
	"flag"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk"
	"github.com/iautre/gowk/auth"
	"github.com/iautre/gowk/auth/proto"
)

func main() {
	flag.Parse()

	// Create servers
	r := gowk.New()
	auth.Router(r.Group(gowk.AuthAPIPrefix()))

	grpcServer := gowk.NewGrpcServer()
	authServer := auth.NewAuthServer()
	proto.RegisterAuthServiceServer(grpcServer.Server, authServer)

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
