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

	// Enable reflection for development
	reflection.Register(s)

	return &GrpcServer{
		Server: s,
	}
}

func (s *GrpcServer) ServerRun() {
	if s.Server == nil {
		s.Server = grpc.NewServer()
	}
	go func() {
		lis, err := net.Listen("tcp", grpcServerAddr)
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port %s: %v", grpcServerAddr, err)
		}

		log.Printf(" [INFO] GrpcServerRun:%s\n", grpcServerAddr)
		if err := s.Server.Serve(lis); err != nil {
			log.Fatalf(" [ERROR] GrpcServerRun:%s err:%v\n", grpcServerAddr, err)
		}
	}()
}

func (s *GrpcServer) ServerStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a channel to signal completion
	done := make(chan struct{})
	go func() {
		s.Server.GracefulStop()
		close(done)
	}()

	// Wait for graceful stop or timeout
	select {
	case <-done:
		log.Printf(" [INFO] GrpcServerStop stopped\n")
	case <-ctx.Done():
		log.Printf(" [INFO] GrpcServerStop timeout\n")
	}
}

// RegisterService registers a gRPC service with the server
func (s *GrpcServer) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.Server.RegisterService(desc, impl)
}
