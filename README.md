# Goark Boot Contrib Arkhos

Official Goark Boot starter for the Arkhos embedded web container.

## Module

- Module: `goark.dev/gbc-arkhos`
- Repository: `github.com/goark-projects/goark-boot-contrib-arkhos`
- License: Apache-2.0
- Default branch: `dev`

## Responsibilities

This module owns the integration between Goark Boot, Arkarta Servlet, and an
Arkhos transport implementation:

- Creates and configures the embedded Servlet container.
- Binds Arkarta deployments to the container.
- Starts and stops the HTTP server through the Boot lifecycle.
- Maps generic `server.*`, Servlet, MVC, and Hertz-specific properties.
- Installs `gbc-log` early and routes Hertz system logs through `slog`.
- Exposes a transport-neutral `Provider` contract for alternative servers.

Hertz is the default provider. A custom provider implements `Provider` and
returns a `ManagedServer`. `ManagedServer.Close` performs immediate termination;
`ManagedServer.Shutdown` performs context-bounded graceful termination.

## Usage

```go
package main

import (
	"context"

	"goark.dev/boot"
	gbcarkhos "goark.dev/gbc-arkhos"
)

func main() {
	app, err := boot.Run(
		context.Background(),
		boot.WithAutoConfiguration(gbcarkhos.AutoConfigure()),
	)
	if err != nil {
		panic(err)
	}
	defer app.Close(context.Background())
}
```

Applications normally import `goark.dev/gbc-web`, which includes this starter.
Use this module directly when supplying Arkarta Servlet deployments manually.

## Configuration

### Generic Server

| Property | Default | Description |
| --- | --- | --- |
| `server.address` | all interfaces | Listen host or IP address |
| `server.port` | `8080` | Listen port; `0` selects an ephemeral port |
| `server.shutdown` | `immediate` | `immediate` or `graceful` |
| `server.max-http-request-header-size` | provider default | Maximum HTTP request header size |

### Hertz

| Property | Default | Description |
| --- | --- | --- |
| `server.hertz.read-timeout` | unset | Full request read timeout |
| `server.hertz.read-header-timeout` | unset | Header timeout used when read timeout is unset |
| `server.hertz.write-timeout` | unset | Response write timeout |
| `server.hertz.idle-timeout` | unset | Keep-alive idle timeout |
| `server.hertz.max-header-bytes` | unset | Hertz-specific header limit |
| `server.hertz.max-request-body-size` | `10MiB` | Complete request body limit |

The generic header limit takes precedence over the Hertz-specific header limit.

### Servlet And MVC

| Property | Default | Description |
| --- | --- | --- |
| `goark.servlet.form.max-body-size` | `10MiB` | URL-encoded form body limit |
| `goark.servlet.multipart.enabled` | `false` | Enables multipart parsing |
| `goark.servlet.multipart.location` | unset | Multipart temporary directory |
| `goark.servlet.multipart.max-file-size` | unset | Single uploaded file limit |
| `goark.servlet.multipart.max-request-size` | unset | Complete multipart request limit |
| `goark.servlet.multipart.file-size-threshold` | unset | In-memory threshold before spilling to disk |
| `goark.mvc.async.request-timeout` | unset | Default Servlet asynchronous request timeout |

Size values are case-insensitive and use binary multiples. Supported suffixes
are `B`, `K`, `KB`, `KiB`, `M`, `MB`, `MiB`, `G`, `GB`, `GiB`, `T`, `TB`,
`TiB`, `P`, `PB`, and `PiB`. `0` retains the default. For maximum-size
properties, `-1` removes the limit; it is not valid for
`goark.servlet.multipart.file-size-threshold`.

```yaml
server:
  address: 127.0.0.1
  port: 8080
  shutdown: graceful
  max-http-request-header-size: 16K
  hertz:
    read-header-timeout: 5s
    max-request-body-size: 20M

goark:
  servlet:
    form:
      max-body-size: 10M
    multipart:
      enabled: true
      max-file-size: 10M
      max-request-size: 20M
  mvc:
    async:
      request-timeout: 30s
```

## Development

```bash
go test ./...
go test -race ./...
go vet ./...
```

## 中文

# Goark Boot Contrib Arkhos（中文）

Goark 官方维护的 Arkhos 嵌入式 Web 容器启动器。

## 职责边界

本模块负责 Goark Boot、Arkarta Servlet 与 Arkhos 传输实现之间的集成：

- 创建并配置嵌入式 Servlet 容器。
- 将 Arkarta Deployment 部署到容器。
- 通过 Boot 生命周期启动和停止 HTTP 服务。
- 映射通用 `server.*`、Servlet、MVC 与 Hertz 专属配置。
- 提前安装 `gbc-log`，并通过 `slog` 接管 Hertz 系统日志。
- 提供与传输实现无关的 `Provider` 扩展契约。

Hertz 是默认 Provider。自定义 Provider 返回 `ManagedServer`：`Close` 负责
立即终止，`Shutdown` 负责受 Context 截止时间约束的优雅关闭。

## 配置

通用服务端配置：

| 属性 | 默认值 | 说明 |
| --- | --- | --- |
| `server.address` | 所有网卡 | 监听主机或 IP 地址 |
| `server.port` | `8080` | 监听端口；`0` 表示随机可用端口 |
| `server.shutdown` | `immediate` | `immediate` 或 `graceful` |
| `server.max-http-request-header-size` | Provider 默认值 | HTTP 请求头最大值 |

Hertz 专属配置：

| 属性 | 默认值 | 说明 |
| --- | --- | --- |
| `server.hertz.read-timeout` | 未设置 | 完整请求读取超时 |
| `server.hertz.read-header-timeout` | 未设置 | 未配置读取超时时使用的请求头超时 |
| `server.hertz.write-timeout` | 未设置 | 响应写出超时 |
| `server.hertz.idle-timeout` | 未设置 | keep-alive 空闲超时 |
| `server.hertz.max-header-bytes` | 未设置 | Hertz 专属请求头限制 |
| `server.hertz.max-request-body-size` | `10MiB` | 完整请求体限制 |

Servlet 与 MVC 配置：

| 属性 | 默认值 | 说明 |
| --- | --- | --- |
| `goark.servlet.form.max-body-size` | `10MiB` | URL 编码表单体限制 |
| `goark.servlet.multipart.enabled` | `false` | 是否启用 multipart 解析 |
| `goark.servlet.multipart.location` | 未设置 | multipart 临时目录 |
| `goark.servlet.multipart.max-file-size` | 未设置 | 单个上传文件限制 |
| `goark.servlet.multipart.max-request-size` | 未设置 | multipart 请求总大小限制 |
| `goark.servlet.multipart.file-size-threshold` | 未设置 | multipart 数据落盘前的内存阈值 |
| `goark.mvc.async.request-timeout` | 未设置 | Servlet 异步请求默认超时 |

大小单位不区分大小写，采用 1024 进制，支持 `B`、`K/KB/KiB`、
`M/MB/MiB`、`G/GB/GiB`、`T/TB/TiB`、`P/PB/PiB`。`0` 表示保留默认值；
最大值类属性可使用 `-1` 表示不限制，但
`goark.servlet.multipart.file-size-threshold` 不接受 `-1`。

## 开发

```bash
go test ./...
go test -race ./...
go vet ./...
```
