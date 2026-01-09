package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk"
	"github.com/iautre/gowk/auth"
)

func main() {
	r := gowk.New()
	auth.Router(r.Group(gowk.AuthAPIPrefix()))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})
	gowk.Run(r)
}
