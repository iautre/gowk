package gowk

import (
	"context"
	"log"
	"log/slog"
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
	engine.NoRoute(func(ctx *gin.Context) {
		ctx.Status(http.StatusNotFound)
		ctx.Abort()
	})
	return engine
}

func (h *HttpServer) ServerRun() {
	if h.Engine == nil {
		h.Engine = gin.Default()
	}
	if h.Handler == nil {
		h.Handler = &http.Server{
			Addr:    httpServerAddr,
			Handler: h.Engine,
			// ReadTimeout:    time.Duration(75) * time.Second,
			// WriteTimeout:   time.Duration(75) * time.Second,
			// MaxHeaderBytes: 1 << uint(20),
		}
	}
	go func() {
		log.Printf(" [INFO] HttpServerRun:%s\n", httpServerAddr)
		if err := h.Handler.ListenAndServe(); err != nil {
			log.Fatalf(" [ERROR] HttpServerRun:%s err:%v\n", httpServerAddr, err)
		}
	}()
}

func (h *HttpServer) ServerStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.Handler.Shutdown(ctx); err != nil {
		log.Fatalf(" [ERROR] HttpServerStop err:%v\n", err)
	}
	log.Printf(" [INFO] HttpServerStop stopped\n")
}
