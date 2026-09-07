# 变更日志

[English](CHANGELOG.md) | 中文

这里记录 Goark Boot Contrib Arkhos 的重要变更。

## [未发布]

暂无未发布变更。

## [0.0.1] - 2026-09-07

### 新增

- Goark Boot 的嵌入式 Arkhos 自动配置。
- 可替换 Hertz 容器 Provider 和显式 Server 属性映射。
- 托管 Server 就绪状态、优雅关闭上下文传递、异步超时、地址和请求限制配置。
- 将 Hertz 日志路由到按组件命名的 `goark.dev/log` Logger。
- 基于 Go 1.26 的跨平台测试、vet 和 race 门禁。

### 变更

- 将所有实际使用的 `golang.org/x` 模块对齐到最新稳定版本。

### 修复

- 嵌入式启动会等待 Server 就绪。
- 关闭过程接收应用上下文，并保持日志生命周期安全。
- Server 地址和限制在启动前完成校验。

[未发布]: https://github.com/goark-projects/goark-boot-contrib-arkhos/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/goark-projects/goark-boot-contrib-arkhos/releases/tag/v0.0.1
