package gowk

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type GrpcServer struct {
	Server *grpc.Server

	health       *health.Server
	healthCancel context.CancelFunc
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

	// 注册标准 grpc.health.v1.Health（必须在 Serve 之前）。整体状态("")由后台 updater
	// 跟随已配置依赖：Postgres/Redis 都连上=SERVING，任一未连上=NOT_SERVING，口径同 HTTP /health。
	s.health = health.NewServer()
	grpc_health_v1.RegisterHealthServer(s.Server, s.health)
	// 先在主线程同步设一次初始状态，避免与下面 Serve 的竞态窗口里被探到 health.NewServer 默认的 SERVING。
	s.refreshHealthStatus()
	healthCtx, cancel := context.WithCancel(context.Background())
	s.healthCancel = cancel
	go s.runHealthUpdater(healthCtx)

	slog.Info("gRPC server running", "addr", lis.Addr().String())
	grpcServerStarted.Store(true)
	go func() {
		if err := s.Server.Serve(lis); err != nil {
			slog.Error("gRPC server serve failed", "addr", lis.Addr().String(), "err", err)
		}
	}()
	return nil
}

// refreshHealthStatus 按"已配置依赖是否就绪"设置 grpc 整体健康状态("")：
// 依赖都连上=SERVING，任一未连上=NOT_SERVING，口径同 HTTP /health。
func (s *GrpcServer) refreshHealthStatus() {
	status := grpc_health_v1.HealthCheckResponse_NOT_SERVING
	if dependenciesReady() {
		status = grpc_health_v1.HealthCheckResponse_SERVING
	}
	s.health.SetServingStatus("", status)
}

// runHealthUpdater 周期性把"已配置依赖是否就绪"同步到 grpc 健康状态。
// 依赖由后台异步连接，运行期状态会变化（连上/断开），故需周期刷新；
// 初始状态已由 ServerRun 在主线程同步设过，这里只负责后续刷新。
func (s *GrpcServer) runHealthUpdater(ctx context.Context) {
	ticker := time.NewTicker(grpcHealthRefreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.refreshHealthStatus()
		}
	}
}

func (s *GrpcServer) ServerStop() {
	// 先停健康状态刷新并 Shutdown：把所有 service 置 NOT_SERVING，让 Watch 中的客户端及时摘流量。
	if s.healthCancel != nil {
		s.healthCancel()
	}
	if s.health != nil {
		s.health.Shutdown()
	}

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
