# Changelog

English | [中文](CHANGELOG.zh-CN.md)

All notable changes to Goark Boot Contrib Arkhos are recorded here.

## [Unreleased]

No unreleased changes.

## [0.0.1] - 2026-09-06

### Added

- Embedded Arkhos auto-configuration for Goark Boot.
- Replaceable Hertz container providers and explicit server property mapping.
- Managed server readiness, graceful shutdown context propagation, async
  timeout, address, and request-limit configuration.
- Hertz logging routed through the component-named `goark.dev/log` logger.
- Cross-platform CI with Go 1.26 tests, vet, and race gates.

### Changed

- Aligned all used `golang.org/x` modules with their latest stable releases.

### Fixed

- Embedded startup waits for server readiness.
- Shutdown receives the application context and logging remains lifecycle-safe.
- Server addresses and limits are validated before startup.

[Unreleased]: https://github.com/goark-projects/goark-boot-contrib-arkhos/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/goark-projects/goark-boot-contrib-arkhos/releases/tag/v0.0.1
