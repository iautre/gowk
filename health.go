package gowk

import (
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// healthCheckArg 是容器 HEALTHCHECK 调用本二进制时使用的子命令：`server healthcheck`。
const healthCheckArg = "healthcheck"

// 各子服务实际启动后置位，供 /health 真实反映组件状态（而非写死）。
// 由 HttpServer.ServerRun / GrpcServer.ServerRun 在监听成功后 Store(true)。
var (
	httpServerStarted atomic.Bool
	grpcServerStarted atomic.Bool
)

// init 在 main() 之前运行：当以 `server healthcheck` 启动时（Docker HEALTHCHECK 用），
// 向本机 /health 发一次存活探测，200 退出 0、否则退出 1，并立即退出进程，
// 不进入业务 main（不连 DB / MQTT，避免探活进程产生副作用）。
//
// 探测端口取自 httpServerAddr（来自 HTTP_SERVER_ADDR 环境变量），与 HTTP server
// 监听用的是同一个变量、同一种取法，Docker 部署里二者一致。
func init() {
	if len(os.Args) > 1 && os.Args[1] == healthCheckArg {
		os.Exit(healthCheckSelf())
	}
}

// healthHandler 是存活探测端点：进程能响应即返回 200。
// 注册在全局中间件之前（见 New），不经过日志/事务中间件，
// 不查 DB/Redis，避免高频探活刷日志、也不在依赖后台重连时误判 unhealthy。
//
// 响应体兼容历史格式 {status,time,services}：
//   - status:   固定 "ok"（探活语义，进程能响应即健康）
//   - time:     当前时间 RFC3339
//   - services: 该服务实际启动的组件，按需出现（http 必有；grpc 仅在启用并启动后出现）。
//     console 这类无 gRPC 的服务只会有 {"http":"running"}，不会出现 grpc。
func healthHandler(c *gin.Context) {
	services := gin.H{}
	if httpServerStarted.Load() {
		services["http"] = "running"
	}
	if grpcServerStarted.Load() {
		services["grpc"] = "running"
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
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
