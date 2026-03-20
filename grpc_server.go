package gowk

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcServer struct {
	Server *grpc.Server
}

func NewGrpcServer() *GrpcServer {
	s := grpc.NewServer()
	reflection.Register(s)
	return &GrpcServer{Server: s}
}

func (s *GrpcServer) ServerRun() {
	if s.Server == nil {
		s.Server = grpc.NewServer()
	}
	go func() {
		lis, err := net.Listen("tcp", grpcServerAddr)
		if err != nil {
			log.Printf(" [ERROR] gRPC 监听失败 %s: %v", grpcServerAddr, err)
			return
		}
		log.Printf(" [INFO] GrpcServerRun:%s\n", grpcServerAddr)
		if err := s.Server.Serve(lis); err != nil {
			log.Printf(" [ERROR] GrpcServerRun:%s err:%v\n", grpcServerAddr, err)
		}
	}()
}

func (s *GrpcServer) ServerStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		s.Server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Printf(" [INFO] GrpcServerStop stopped\n")
	case <-ctx.Done():
		log.Printf(" [WARN] GrpcServerStop timeout，强制停止\n")
		s.Server.Stop()
	}
}

func (s *GrpcServer) RegisterService(desc *grpc.ServiceDesc, impl any) {
	s.Server.RegisterService(desc, impl)
}
