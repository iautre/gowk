package gowk

import (
	"log"
	"net"

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
	s.Server.GracefulStop()
	log.Printf(" [INFO] GrpcServerStop stopped\n")
}

// RegisterService registers a gRPC service with the server
func (s *GrpcServer) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.Server.RegisterService(desc, impl)
}
