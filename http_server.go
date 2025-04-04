package gowk

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk/conf"
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

func Run(r *gin.Engine) {
	httpServer := &HttpServer{
		Engine: r,
	}
	httpServer.Run()
}

func (h *HttpServer) Run() {
	if h.Engine == nil {
		h.Engine = gin.Default()
	}
	if h.Handler == nil {
		h.Handler = &http.Server{
			Addr:    conf.Server().Addr,
			Handler: h.Engine,
			// ReadTimeout:    time.Duration(75) * time.Second,
			// WriteTimeout:   time.Duration(75) * time.Second,
			// MaxHeaderBytes: 1 << uint(20),
		}
	}

	h.ServerRun()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	h.ServerStop()
}

func (h *HttpServer) ServerRun() {
	go func() {
		if conf.Server().Cert == "" {
			log.Printf(" [INFO] HttpServerRun:%s\n", conf.Server().Addr)
			if err := h.Handler.ListenAndServe(); err != nil {
				log.Fatalf(" [ERROR] HttpServerRun:%s err:%v\n", conf.Server().Addr, err)
			}
		} else {
			log.Printf(" [INFO] Http2ServerRun:%s\n", conf.Server().Addr)
			if err := h.Handler.ListenAndServeTLS(conf.Server().Cert, conf.Server().Key); err != nil {
				log.Fatalf(" [ERROR] Http2ServerRun:%s err:%v\n", conf.Server().Addr, err)
			}
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
