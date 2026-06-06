package gowk

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// healthCheckArg 是容器 HEALTHCHECK 调用本二进制时使用的 HTTP 探活子命令：`server healthcheck`。
const healthCheckArg = "healthcheck"

// grpcHealthCheckArg 是纯 gRPC 容器 HEALTHCHECK 复用本二进制做 gRPC 探活的子命令：`server grpc-healthcheck`。
const grpcHealthCheckArg = "grpc-healthcheck"

// 各子服务实际启动后置位，供 /health 真实反映组件状态（而非写死）。
// 由 HttpServer.ServerRun / GrpcServer.ServerRun 在监听成功后 Store(true)。
var (
	httpServerStarted atomic.Bool
	grpcServerStarted atomic.Bool
)

// init 在 main() 之前运行，分发容器 HEALTHCHECK 用的探活子命令，探完立即退出进程，
// 不进入业务 main（不连 DB/MQTT，避免探活进程产生副作用）：
//   - `server healthcheck`：向本机 HTTP /health 发探测，200 退出 0、否则退出 1；
//   - `server grpc-healthcheck`：向本机 gRPC 端口发标准 grpc.health.v1.Health/Check，SERVING 退出 0、否则退出 1。
//
// 探测端口分别取自 httpServerAddr / grpcServerAddr（与 server 监听用的是同一组环境变量）。
func init() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case healthCheckArg:
			os.Exit(healthCheckSelf())
		case grpcHealthCheckArg:
			os.Exit(grpcHealthCheckSelf())
		}
	}
}

// dependenciesReady 报告所有"已配置"的关键依赖是否都已连上，口径与 /health 一致：
//   - 配了 DATABASE_DSN 则 Postgres 必须已连上；
//   - 配了 REDIS_ADDR 则 Redis 必须已连上；
//   - 未配置的依赖不参与判定。
//
// 供 HTTP /health 与 gRPC grpc.health.v1.Health 共用，保证两条链路判定一致。
func dependenciesReady() bool {
	if databaseDsn != "" && defaultPostgres.Load() == nil {
		return false
	}
	if HasRedis() && defaultRedis.Load() == nil {
		return false
	}
	return true
}

// healthHandler 是健康检查端点：进程存活且已配置的关键依赖（Postgres/Redis）都连上才返回 200。
// 注册在全局中间件之前（见 New），不经过日志/事务中间件。
// 依赖状态只读 atomic 指针（defaultPostgres/defaultRedis），不发起真实查询、不刷日志。
//
// 判定规则：
//   - 配置了 DATABASE_DSN：Postgres 连上前（后台仍在重试）一律视为未连上 → 不健康。
//   - 配置了 REDIS_ADDR：Redis 连上前一律视为未连上 → 不健康。
//   - 未配置的依赖不参与判定。
//
// 响应体 {status,time,services}：
//   - status:   全部就绪为 "ok"，否则 "unavailable"。
//   - time:     当前时间 RFC3339。
//   - services: 各组件状态。http/grpc 监听成功为 "running"；
//     postgres/redis 为 "connected"（已连上）或 "down"（已配置但未连上）。
//
// HTTP 状态码：健康 200，不健康 503，供 Docker HEALTHCHECK / K8s readiness 判定。
func healthHandler(c *gin.Context) {
	services := gin.H{}
	if httpServerStarted.Load() {
		services["http"] = "running"
	}
	if grpcServerStarted.Load() {
		services["grpc"] = "running"
	}

	healthy := true

	if databaseDsn != "" {
		if defaultPostgres.Load() != nil {
			services["postgres"] = "connected"
		} else {
			services["postgres"] = "down"
			healthy = false
		}
	}

	if HasRedis() {
		if defaultRedis.Load() != nil {
			services["redis"] = "connected"
		} else {
			services["redis"] = "down"
			healthy = false
		}
	}

	status := "ok"
	code := http.StatusOK
	if !healthy {
		status = "unavailable"
		code = http.StatusServiceUnavailable
	}

	c.JSON(code, gin.H{
		"status":   status,
		"time":     time.Now().Format(time.RFC3339),
		"services": services,
	})
}

// healthCheckSelf 向本进程监听的 /health 发请求，成功(200)返回 0，否则返回 1。
// 端口取自 httpServerAddr（来自 HTTP_SERVER_ADDR 环境变量），与服务监听端口一致。
func healthCheckSelf() int {
	addr := httpServerAddr
	if strings.HasPrefix(addr, ":") {
		addr = "127.0.0.1" + addr
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://" + addr + "/health")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return 0
	}
	return 1
}

// grpcHealthCheckSelf 向本进程 gRPC 端口发一次标准 grpc.health.v1.Health/Check，
// SERVING 返回 0，否则返回 1。供纯 gRPC（scratch 无 shell/grpc_health_probe）容器在
// HEALTHCHECK 里复用本二进制自探：CMD ["/server", "grpc-healthcheck"]。
// 端口取自 grpcServerAddr（GRPC_SERVER_ADDR），与 gRPC server 监听端口一致；未配置直接判失败。
func grpcHealthCheckSelf() int {
	addr := grpcServerAddr
	if addr == "" {
		return 1
	}
	if strings.HasPrefix(addr, ":") {
		addr = "127.0.0.1" + addr
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return 1
	}
	defer conn.Close()
	resp, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return 1
	}
	if resp.GetStatus() == grpc_health_v1.HealthCheckResponse_SERVING {
		return 0
	}
	return 1
}
