# Goark Boot Contrib Arkhos

Official Goark Boot starter module for the Arkhos embedded web container.

## Module

- Module path: `goark.dev/gbc-arkhos`
- Repository: `github.com/goark-projects/goark-boot-contrib-arkhos`
- License: Apache-2.0
- Default branch: `main`
- Development branch: `dev`

## Scope

`goark.dev/gbc-arkhos` is the Goark-managed Arkhos starter. It is intended to provide the Spring Boot embedded Tomcat starter equivalent for the Goark ecosystem:

- Arkhos embedded container bootstrap.
- Arkarta servlet deployment binding.
- Boot lifecycle integration for start and graceful stop.
- Server configuration mapping for address, timeouts, header limits, multipart parsing, and Servlet async timeout.
- A transport-neutral provider boundary based on `arkarta/servlet/container.Container`.

Hertz is the default embedded implementation. `net/http` remains an explicit Arkhos reference and compatibility implementation; it is not selected implicitly by this starter.

## Container replacement

Application deployments depend only on Arkarta contracts. A future gnet or fasthttp starter must expose the same call shape:

```go
boot.WithAutoConfiguration(containerstarter.AutoConfigure())
```

Replacing the container changes the starter module/import only. Business handlers, Arkarta deployments, configuration property names, and the `AutoConfigure()` call remain unchanged. Go cannot activate a module that is only present in `go.mod` and never imported, so dependency-only runtime discovery is intentionally not used.

## Usage

```go
package main

import (
	"context"

	"goark.dev/boot"
	"goark.dev/gbc-arkhos"
)

func main() {
	app, err := boot.Run(context.Background(),
		boot.WithAutoConfiguration(gbcarkhos.AutoConfigure()),
	)
	if err != nil {
		panic(err)
	}
	defer app.Close(context.Background())
}
```

Application code usually uses `goark.dev/gbc-web`, which includes this starter by default.
Use this module directly when you need the embedded Arkhos container with manually supplied Arkarta Servlet deployments.

## Configuration Properties

| Property | Default | Description |
| --- | --- | --- |
| `goark.web.server.address` | `:8080` | TCP listen address passed to Arkhos. |
| `goark.web.server.read-timeout` | unset | Full request read timeout, parsed as Go duration. |
| `goark.web.server.read-header-timeout` | unset | Request header read timeout, parsed as Go duration. |
| `goark.web.server.write-timeout` | unset | Response write timeout, parsed as Go duration. |
| `goark.web.server.idle-timeout` | unset | Keep-alive idle timeout, parsed as Go duration. |
| `goark.web.server.max-header-bytes` | unset | Maximum HTTP header bytes. |
| `goark.web.server.max-request-body-size` | `10MiB` | Maximum complete HTTP request body size. |
| `goark.web.servlet.form.max-body-size` | `10MiB` | Maximum URL-encoded form body size. |
| `goark.web.servlet.multipart.location` | unset | Temporary directory for multipart files. |
| `goark.web.servlet.multipart.max-file-size` | unset | Maximum single uploaded file size in bytes. |
| `goark.web.servlet.multipart.max-request-size` | unset | Maximum multipart request size in bytes. |
| `goark.web.servlet.multipart.file-size-threshold` | unset | In-memory threshold before multipart data spills to disk. |
| `goark.web.servlet.async.timeout` | unset | Default Servlet async timeout, parsed as Go duration. |

`spring.mvc.async.request-timeout` is accepted as a Spring-compatible alias for `goark.web.servlet.async.timeout`.

## Development

```bash
go test ./...
```

## Chinese

# Goark Boot Contrib Arkhos（中文）

Goark 官方维护的 Arkhos 嵌入式 Web 容器启动器模块。

## 模块信息

- 模块路径：`goark.dev/gbc-arkhos`
- 仓库地址：`github.com/goark-projects/goark-boot-contrib-arkhos`
- 开源协议：Apache-2.0
- 默认分支：`main`
- 开发分支：`dev`

## 职责边界

`goark.dev/gbc-arkhos` 是 Goark 生态中对标 Spring Boot 嵌入式 Tomcat 启动器的官方模块：

- 启动 Arkhos 嵌入式容器。
- 绑定 Arkarta Servlet 部署模型。
- 接入 Boot 生命周期，支持启动与优雅停止。
- 映射地址、超时、请求头限制、multipart 和 Servlet async 超时等服务端配置。
- 基于 `arkarta/servlet/container.Container` 提供与传输实现无关的 Provider 边界。

Hertz 是默认嵌入式实现。`net/http` 保留为 Arkhos 显式参考实现和兼容实现，本 starter 不会隐式选择它。

## 容器替换

业务 Deployment 只依赖 Arkarta 标准契约。未来 gnet 或 fasthttp starter 必须提供相同调用形态：

```go
boot.WithAutoConfiguration(containerstarter.AutoConfigure())
```

替换容器时只修改 starter 模块依赖和 import；业务 Handler、Arkarta Deployment、配置属性名以及 `AutoConfigure()` 调用均保持不变。Go 不会链接仅写入 `go.mod` 而未被 import 的模块，因此不采用不可靠的纯依赖运行时发现机制。

## 使用方式

业务应用通常直接使用 `goark.dev/gbc-web`，它会默认包含 Arkhos 嵌入式容器。
只有在需要手工提供 Arkarta Servlet 部署对象时，才直接使用本模块。

## 配置属性

| 属性 | 默认值 | 说明 |
| --- | --- | --- |
| `goark.web.server.address` | `:8080` | Arkhos TCP 监听地址。 |
| `goark.web.server.read-timeout` | 未设置 | 完整请求读取超时，按 Go duration 解析。 |
| `goark.web.server.read-header-timeout` | 未设置 | 请求头读取超时，按 Go duration 解析。 |
| `goark.web.server.write-timeout` | 未设置 | 响应写出超时，按 Go duration 解析。 |
| `goark.web.server.idle-timeout` | 未设置 | keep-alive 空闲超时，按 Go duration 解析。 |
| `goark.web.server.max-header-bytes` | 未设置 | HTTP 请求头最大字节数。 |
| `goark.web.server.max-request-body-size` | `10MiB` | 完整 HTTP 请求体最大值。 |
| `goark.web.servlet.form.max-body-size` | `10MiB` | URL 编码表单请求体最大值。 |
| `goark.web.servlet.multipart.location` | 未设置 | multipart 临时文件目录。 |
| `goark.web.servlet.multipart.max-file-size` | 未设置 | 单个上传文件最大字节数。 |
| `goark.web.servlet.multipart.max-request-size` | 未设置 | multipart 请求最大字节数。 |
| `goark.web.servlet.multipart.file-size-threshold` | 未设置 | multipart 数据落盘前的内存阈值。 |
| `goark.web.servlet.async.timeout` | 未设置 | Servlet async 默认超时，按 Go duration 解析。 |

`spring.mvc.async.request-timeout` 可作为 `goark.web.servlet.async.timeout` 的 Spring 兼容别名。

## 开发

```bash
go test ./...
```
