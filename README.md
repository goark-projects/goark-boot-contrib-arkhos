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
- Server configuration mapping for address, timeouts, context path, multipart, sessions, and optional profiles.

The initial repository bootstrap exposes stable module metadata only. Runtime auto-configuration APIs will be added in later implementation slices.

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
- 映射地址、超时、上下文路径、multipart、session 和可选 profile 等服务端配置。

当前初始化版本只暴露稳定的模块元数据。运行期自动配置 API 会在后续实现切片中补齐。

## 开发

```bash
go test ./...
```
