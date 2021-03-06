package gowk

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	Handler *http.Server
	Engine  *gin.Engine
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
			Addr:           Conf().GetServer().Addr,
			Handler:        h.Engine,
			ReadTimeout:    time.Duration(10) * time.Second,
			WriteTimeout:   time.Duration(10) * time.Second,
			MaxHeaderBytes: 1 << uint(20),
		}
	}

	h.ServerRun()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	h.ServerStop()
}

func (h *HttpServer) ServerRun() {
	go func() {
		log.Printf(" [INFO] HttpServerRun:%s\n", Conf().GetServer().Addr)
		if err := h.Handler.ListenAndServe(); err != nil {
			log.Fatalf(" [ERROR] HttpServerRun:%s err:%v\n", Conf().GetServer().Addr, err)
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
