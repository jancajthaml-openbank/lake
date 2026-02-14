# Constraints — Lake

## Constraints table

| Dimension | Value | Source |
|-----------|-------|--------|
| Network model | Always-on (localhost and LAN TCP) | Codebase: ZMQ binds `0.0.0.0`, StatsD sends to `127.0.0.1`, all communication is TCP/UDP on a single host or local network |
| Operator profile | Minimal | Codebase: 4 environment variables, systemd `Restart=always`, no manual recovery procedures, no unseal ceremony, no cluster coordination. The operator installs a `.deb` and starts a systemd unit. |
| Hardware profile | Commodity (single host, modest resources) | README: benchmarked at 2 GB RAM, 1 CPU. Systemd unit sets `LimitNOFILE=1048576`. No multi-node deployment. |
| Component budget | 2 — the lake binary + libzmq5 shared library | Derived: operator profile is minimal. The current system has exactly 2 runtime components (the binary, the C library). The conversion must not increase this. StatsD daemon is external and pre-existing. |
| Internet access | Not required at runtime | Codebase: no outbound HTTP, no cloud SDK, no package registry access at runtime. Internet only needed at build time (Zig compiler download, libzmq headers). |
| Target OS | Linux (Debian-family, systemd) | Codebase: systemd units, `/proc` filesystem access, `sd_notify` protocol, `.deb` packaging. No Windows, no macOS at runtime. |
| Target architectures | amd64, arm64 | Codebase: two Dockerfiles, two `.install` files, CI matrix builds both. |
| Compliance | None stated | Assumed — confirm with user. No PCI, HIPAA, SOC2, or data sovereignty requirements visible in codebase. ZMQ uses NULL mechanism (no encryption). |
| Zig compiler maturity | Zig is pre-1.0 (latest stable is 0.13.x as of 2025) | Prompt: user chose Zig. Risk: breaking changes between Zig versions. Constraint: pin a specific Zig version in the build system and CI. |
| Performance baseline | ≥ 750,000 messages/sec on commodity hardware (1 CPU, 2 GB RAM) | README + perf tests. The conversion must not regress below this baseline. |
| Binary size | Small (stripped Rust binary is a few MB) | Codebase: `objcopy --strip-unneeded`. Constraint: Zig binary must be comparable or smaller. |
| Startup time | < 1 second | Codebase: `TimeoutStartSec=1` in systemd unit. The binary must reach `READY=1` within 1 second. |
| Shutdown time | < 3 seconds | Codebase: `TimeoutStopSec=3` in systemd unit. The binary must reach exit within 3 seconds of SIGTERM. |
| Existing test suite compatibility | Blackbox tests must pass unchanged | Codebase: Python behave tests exercise the binary externally via ZMQ, systemd, and StatsD. They are language-agnostic. The converted binary must pass all 6 feature files without test modification. |

## Invariants

1. The converted binary must produce identical ZMTP 3.0 wire behaviour — any ZMQ client that connects to the current Rust binary must connect identically to the Zig binary with no client-side changes.
2. The converted binary must emit identical StatsD metric names (`openbank.lake.message.relayed`, `openbank.lake.memory.bytes`) with the same StatsD protocol encoding.
3. The converted binary must send `READY=1` and `STOPPING=1` to `$NOTIFY_SOCKET` using the sd_notify Unix datagram protocol, or systemd will kill it.
4. The converted binary must exit cleanly within 3 seconds of receiving SIGTERM, or systemd will SIGKILL it and the operator will see unclean shutdowns in logs.
5. The converted binary must be installable as a `.deb` package with the same package name (`lake`), the same binary path (`/usr/bin/lake`), and the same systemd unit names.
6. All 6 existing blackbox test features must pass against the Zig binary without any test modifications.
7. Throughput must not regress below 750,000 messages/sec on equivalent hardware (1 CPU, 2 GB RAM, amd64).
8. The runtime dependency on `libzmq5 >= 4.3, < 4.4` must be preserved or explicitly replaced (static linking is acceptable but must be decided, not accidental).
9. No additional runtime infrastructure components may be introduced — the operator manages 1 binary and 1 shared library; the conversion must not add a second daemon, a config database, a secrets manager, or any other component.
10. The Zig source must be buildable for both amd64 and arm64 Linux from a single `build.zig`, as the current Cargo setup builds both targets from a single `Cargo.toml`.

## Failure scenarios

| # | Scenario | What fails | Blast radius | Expected recovery |
|---|----------|-----------|--------------|-------------------|
| 1 | ZMQ PULL socket bind failure on startup | `zmq_bind` returns error (port in use, permissions) | Total — no messages relayed | Operator checks port conflict or permissions. systemd `Restart=always` will retry. Binary logs the error and exits. Current Rust binary does the same. |
| 2 | ZMQ PUB socket bind failure on startup | `zmq_bind` returns error | Total — no messages relayed | Same as scenario 1. |
| 3 | StatsD endpoint unreachable | UDP send fails silently or StatsD daemon is down | Metrics lost; relay continues operating. No message impact. | Operator restarts StatsD daemon. Metrics resume. No data loss because the relay is stateless. The `statsd` crate (and the Zig replacement) must not block or crash on UDP send failure. |
| 4 | SIGTERM during active relay | Main thread receives SIGTERM while relay loop is blocking on `zmq_msg_recv` | Graceful — `zmq_ctx_shutdown` (per ADR-02-02) causes `zmq_msg_recv` to return `ETERM`, breaking the loop. Threads must join within 3 seconds. | No operator action. systemd logs clean shutdown. If the join hangs past 3 seconds, systemd sends SIGKILL. ADR-02-02 replaced the fragile self-PUSH trick with `zmq_ctx_shutdown` for deterministic shutdown. |
| 5 | libzmq5 not installed or wrong version | Binary fails to load shared library at process start | Total — binary does not start | Operator installs `libzmq5` per `.deb` dependency. systemd restart will succeed once library is present. Alternative: Zig build statically links libzmq, eliminating this failure mode entirely. |
| 6 | `/etc/lake/conf.d/init.conf` missing at boot | `lake.service` (parent) has `ConditionPathExists=/etc/lake/conf.d/init.conf` — service does not start | Total — relay does not start | Operator creates the config file. Package install creates it by default. |
| 7 | Zig compiler version mismatch in CI | Zig is pre-1.0; a CI image upgrade may introduce breaking language changes | Build fails — no binary produced | Pin Zig version in `build.zig` and CI Dockerfile. Do not use `latest` tag. |
| 8 | Performance regression after conversion | Zig binary achieves < 750K msg/sec | Functional but violated performance contract | Run perf test suite as part of CI gate. If regression detected, investigate: is it the relay loop (C call overhead), thread coordination (atomic operations), or an allocation in the hot path? |
| 9 | Resource leak in Zig code (no borrow checker) | Missing `defer` or `deinit` call causes file descriptor / memory leak | Gradual — memory or fd exhaustion over hours/days. Relay eventually crashes. systemd restarts it. | Code review discipline. Run blackbox tests with extended duration. Monitor `openbank.lake.memory.bytes` for monotonic growth. Zig's `GeneralPurposeAllocator` in debug mode detects leaks — use it in test builds. |
| 10 | Thread join deadlock on shutdown | Relay or metrics thread does not exit within 3 seconds; main thread blocks on join | systemd sends SIGKILL after `TimeoutStopSec=3`. Unclean shutdown. | The current Rust code has this risk too (the shutdown-via-PUSH pattern is a mitigation). The Zig conversion should preserve or improve this pattern. Consider using `zmq_ctx_shutdown` instead of the self-PUSH trick for cleaner unblocking. |
