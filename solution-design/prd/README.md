# Product Decision Records

| # | Title | Summary |
|---|-------|---------|
| PRD-001 | [High-throughput message relay](prd-001-high-throughput-message-relay.md) | Relay messages PULL→PUB at ≥750K msg/sec with zero per-message heap allocation, preserved ordering, and backpressure handling |
| PRD-002 | [Operational observability](prd-002-operational-observability.md) | Emit StatsD metrics (throughput count, memory gauge) and structured log output to stdout without impacting relay performance |
| PRD-003 | [Daemon lifecycle and host integration](prd-003-daemon-lifecycle.md) | Start within 1s (sd_notify READY), shut down within 3s of SIGTERM, clean resource release, env-based configuration |
| PRD-004 | [Deployment and distribution](prd-004-deployment-and-distribution.md) | .deb package, Docker images, dual-arch (amd64/arm64), no new runtime components, identical operator interface |
| PRD-005 | [Conversion contract preservation](prd-005-conversion-contract-preservation.md) | All 13 external contracts (wire protocol, metrics, sd_notify, env vars, packaging) identical; all 6 blackbox tests pass unchanged |
| PRD-006 | [Build and compilation infrastructure](prd-006-build-and-compilation.md) | Zig build system, pinned compiler version, dual-arch cross-compilation, CI pipeline integration, dynamic/static linking option |
