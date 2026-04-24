package gowk

import (
	"context"
	"fmt"
	"log/slog"
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

// ServerRun 同步绑定端口并在后台 Serve。
// GRPC_SERVER_ADDR 未配置视为"未启用"，返回 nil（不是错误）；
// 已配置但监听失败，返回 error 交由调用方 fail-fast；
// Serve 阶段的错误只打日志。
func (s *GrpcServer) ServerRun() error {
	if !HasGRPC() {
		slog.Info("GRPC_SERVER_ADDR 未配置，跳过 gRPC 启动")
		return nil
	}
	if s.Server == nil {
		s.Server = grpc.NewServer()
	}
	lis, err := net.Listen("tcp", grpcServerAddr)
	if err != nil {
		return fmt.Errorf("gRPC 监听失败 addr=%s: %w", grpcServerAddr, err)
	}
	slog.Info("gRPC server running", "addr", lis.Addr().String())
	go func() {
		if err := s.Server.Serve(lis); err != nil {
			slog.Error("gRPC server serve failed", "addr", lis.Addr().String(), "err", err)
		}
	}()
	return nil
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
		slog.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		slog.Warn("gRPC server graceful stop timeout, forcing stop")
		s.Server.Stop()
	}
}

func (s *GrpcServer) RegisterService(desc *grpc.ServiceDesc, impl any) {
	s.Server.RegisterService(desc, impl)
}
