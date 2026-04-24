这是一个gin框架的公共模块

## 环境变量与依赖语义

- `DATABASE_DSN`：配置后 `gowk.Run` / `InitPostgres` 会触发 Postgres 后台初始化（非阻塞）。连不上**不会退出进程**，后台以指数退避无限重试。未就绪期间 `gowk.Postgres(ctx)` 返回 nil，`gowk.PostgresTx(ctx)` 返回 `postgres unavailable` 错误，不会 panic。未配置或 DSN 解析失败同样保持降级。
- `REDIS_ADDR`：配置后首次 `gowk.Redis()` / `InitRedis()` 触发 Redis 后台初始化（非阻塞）。连不上同样不退出进程，后台以指数退避无限重试。未就绪期间 `gowk.Redis()` 返回 nil；未配置也返回 nil。
- 进程收到 SIGINT / SIGTERM 时 `closePostgres` / `closeRedis` 会取消后台重试 goroutine 并关闭已就绪的连接。

## 后台重试策略

- 退避：初始 `base`，每轮 × 2，封顶 `max`；命中 `max` 之后保持 `max` 间隔无限重试。
- 进程退出（SIGINT/SIGTERM）时通过 `context.Cancel` 立即打断当前 sleep，goroutine 尽快退出。
- 每轮失败日志带 `attempt` / `elapsed` / `backoff` / `err`；连接成功打印 `attempts` / `elapsed`；取消时打印 `已取消` 与已尝试次数。
- 运行期断线由底层客户端（pgxpool / go-redis）自行透明重连，不走本模块的后台重试。

### 可配置环境变量

时长格式走 Go `time.ParseDuration`（`2s`、`500ms`、`1m`、`30s` 等）；未设置、解析失败或 `<= 0` 时回落默认值。

| 变量 | 默认 | 含义 |
|---|---|---|
| `DATABASE_RETRY_BASE_INTERVAL` | `2s` | Postgres 后台重试初始退避 |
| `DATABASE_RETRY_MAX_INTERVAL` | `30s` | Postgres 后台重试退避封顶 |
| `DATABASE_PING_TIMEOUT` | `5s` | Postgres 单次 `NewWithConfig` + `Ping` 超时 |
| `REDIS_RETRY_BASE_INTERVAL` | `2s` | Redis 后台重试初始退避 |
| `REDIS_RETRY_MAX_INTERVAL` | `30s` | Redis 后台重试退避封顶 |
| `REDIS_PING_TIMEOUT` | `5s` | Redis 单次 `Ping` 超时 |

## HTTP / gRPC 启动语义（fail-fast）

配置即意图：填了地址就视作必须可用。

- HTTP 总是启用（`HTTP_SERVER_ADDR` 默认 `:3030`）。`net.Listen` 绑定端口成功后才打印 `HTTP server running`，绑定失败 `slog.Error + os.Exit(1)`；随后 `Serve` 在 goroutine 内运行，非 `http.ErrServerClosed` 的错误只走 `HTTP server serve failed` 日志。
- gRPC 由 `GRPC_SERVER_ADDR` 决定是否启用：未配置 → 安静跳过；配置了 → 监听成功才打印 `gRPC server running`，绑定失败同样 `slog.Error + os.Exit(1)`。
- `Run()` 在 fail-fast 时会先做一次尽力而为的清理（已起的 HTTP `Shutdown`、`closePostgres` / `closeRedis` 取消后台重试），再 `os.Exit(1)`，交给 `docker restart: always` / K8s `restartPolicy: Always` 重启。
- 打印的 `addr` 取自 `ln.Addr().String()`，因此绑定 `:0` 这类系统分配端口时日志里是实际端口。
