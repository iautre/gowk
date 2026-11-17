这是一个gin框架的公共模块

## MCP Streamable HTTP

`SetupMCP` / `MCPHandler` 提供无状态的 MCP Streamable HTTP 服务。

需要自定义传输选项（例如关闭 DNS rebinding 保护）时用 `SetupMCPWithOptions` / `MCPHandlerWithOptions`。

### 同机反向代理会 403

MCP SDK 自 v1.8.0 起默认开启 DNS rebinding 保护：**连接落在回环地址、但 `Host` 头不是回环地址**的请求一律 403。

所以如果服务挂在同机 nginx 后面、且代理透传了真实域名：

```nginx
proxy_set_header Host $host;   # 会触发 403
```

二选一解决：

- 让代理改写 Host：`proxy_set_header Host 127.0.0.1:8080;`
- 或在代码里关掉保护：

```go
gowk.SetupMCPWithOptions(router, "/mcp", provider, &mcp.StreamableHTTPOptions{
    Stateless:                  true,
    DisableLocalhostProtection: true,
}, mw.Login)
```

注意自定义 `opts` 会整体替换默认值——默认的 `Stateless` / `PropagateRequestCancellation` 需要自己带上。


## 环境变量与依赖语义

- `DATABASE_ENABLED`：默认 `true`（严格模式）。启用时 `gowk.Run` 会**同步**建池并 `Ping`，`DATABASE_DSN` 为空或 `DATABASE_MIGRATION_CONNECT_TIMEOUT` 内连不上就 `slog.Error + os.Exit(1)`。只有显式设为 `false` 才允许在没有 PostgreSQL 的情况下运行，此时 `gowk.Postgres(ctx)` 恒为 nil，配了 `DATABASE_DSN` 也会被忽略并记一条 warn。
- `DATABASE_DSN`：PostgreSQL 连接串。启用时必须配置；`InitPostgres()` 返回 nil 之后 `gowk.Postgres(ctx)` 必定非 nil，`PostgresTx` 不会因未就绪而报 `postgres unavailable`。
- `REDIS_ADDR`：配置后首次 `gowk.Redis()` / `InitRedis()` 触发 Redis 后台初始化（非阻塞）。连不上**不退出进程**，后台以指数退避无限重试。未就绪期间 `gowk.Redis()` 返回 nil；未配置也返回 nil。
- 进程收到 SIGINT / SIGTERM 时 `closePostgres` 关闭已就绪的连接池，`closeRedis` 除关闭客户端外还会取消 Redis 的后台重试 goroutine。

## 依赖就绪策略

PostgreSQL 与 Redis 的语义不同，这是刻意的：

- **PostgreSQL：启动期有界重试，失败即退出。** 数据库是业务数据源，缺了服务不该对外提供错误结果。启动时在 `DATABASE_MIGRATION_CONNECT_TIMEOUT`（默认 30s）内按 `DATABASE_RETRY_BASE_INTERVAL` → `DATABASE_RETRY_MAX_INTERVAL` 退避重试，超时仍未连上则返回错误、进程退出，交给容器重启策略自愈。
- **Redis：后台无限重试。** 视为可降级的缓存/会话依赖，连不上不阻塞启动。
- 运行期断线由底层客户端（pgxpool / go-redis）自行透明重连，不走本模块的重试。

### 可配置环境变量

时长格式走 Go `time.ParseDuration`（`2s`、`500ms`、`1m`、`30s` 等）；未设置、解析失败或 `<= 0` 时回落默认值。布尔变量走 `strconv.ParseBool`（`1/0`、`t/f`、`true/false`），未设置或无法解析时用默认值。

| 变量 | 默认 | 含义 |
|---|---|---|
| `DATABASE_ENABLED` | `true` | 是否启用 PostgreSQL；`false` 时跳过连接、`Postgres(ctx)` 恒 nil |
| `DATABASE_RETRY_BASE_INTERVAL` | `2s` | Postgres 启动期重试初始退避 |
| `DATABASE_RETRY_MAX_INTERVAL` | `30s` | Postgres 启动期重试退避封顶 |
| `DATABASE_PING_TIMEOUT` | `5s` | Postgres 单次 `NewWithConfig` + `Ping` 超时 |
| `DATABASE_MIGRATION_CONNECT_TIMEOUT` | `30s` | 启动期建池与迁移连接的总超时，超时即 fail-fast |
| `REDIS_RETRY_BASE_INTERVAL` | `2s` | Redis 后台重试初始退避 |
| `REDIS_RETRY_MAX_INTERVAL` | `30s` | Redis 后台重试退避封顶 |
| `REDIS_PING_TIMEOUT` | `5s` | Redis 单次 `Ping` 超时 |

## HTTP / gRPC 启动语义（fail-fast）

配置即意图：填了地址就视作必须可用。

- HTTP 总是启用（`HTTP_SERVER_ADDR` 默认 `:3030`）。`net.Listen` 绑定端口成功后才打印 `HTTP server running`，绑定失败 `slog.Error + os.Exit(1)`；随后 `Serve` 在 goroutine 内运行，非 `http.ErrServerClosed` 的错误只走 `HTTP server serve failed` 日志。
- gRPC 由 `GRPC_SERVER_ADDR` 决定是否启用：未配置 → 安静跳过；配置了 → 监听成功才打印 `gRPC server running`，绑定失败同样 `slog.Error + os.Exit(1)`。
- `Run()` 在 fail-fast 时会先做一次尽力而为的清理（已起的 HTTP `Shutdown`、`closePostgres` / `closeRedis`），再 `os.Exit(1)`，交给 `docker restart: always` / K8s `restartPolicy: Always` 重启。
- 打印的 `addr` 取自 `ln.Addr().String()`，因此绑定 `:0` 这类系统分配端口时日志里是实际端口。
