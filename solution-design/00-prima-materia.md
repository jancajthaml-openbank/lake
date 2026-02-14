# Prima Materia — Lake

Lake is a high-performance ZMQ message relay for the OpenBank platform, currently implemented in Rust (edition 2021) as a single binary crate (~400 lines, 9 modules). It receives messages on a ZMQ PULL socket and fans them out via a ZMQ PUB socket over ZMTP 3.0/TCP, enabling inter-service communication at ~750K messages/sec. The target is a full conversion from Rust to Zig, retaining the identical ZMQ wire protocol, systemd integration, and packaging contracts while improving reliability (eliminating unsafe FFI indirection, strengthening error handling) and performance (reducing allocation overhead, leveraging Zig's comptime and direct C interop).

## Domain capabilities

| Domain | Purpose | Cloud touchpoints |
|--------|---------|-------------------|
| Message relay | Receives messages from upstream services via ZMQ PULL and fans them out to downstream subscribers via ZMQ PUB, providing a decoupled communication bus for the OpenBank platform | None |
| Metrics reporting | Periodically reports message throughput count and process memory usage (RSS) to a StatsD daemon over UDP for operational observability | None |
| Process lifecycle | Manages daemon startup and graceful shutdown with systemd sd_notify integration (READY/STOPPING) and POSIX signal handling (SIGTERM) | None |
| Configuration | Loads runtime parameters (ports, log level, StatsD endpoint) from environment variables with sensible defaults; config file watched by systemd for restart-on-change | None |
| Logging | Structured timestamped log output with ISO-8601 UTC timestamps and colored level indicators to stdout for journald capture | None |
| Error handling | Maps ZMQ C library errno codes to named error variants with human-readable messages for diagnostics | None |

## Current → desired mapping

| Capability | Current | Desired | ADR | Challenged |
|-----------|---------|---------|-----|-----------|
| Language / runtime | Rust 2021 (compiled, no runtime) | Zig (compiled, no runtime) | [ADR-01-01](adr/01-platform/adr-01-01-language-zig.md) | necessary — user-directed conversion target |
| Build system | Cargo (single crate, Cargo.toml/lock) | Zig build system (`build.zig`) | [ADR-01-02](adr/01-platform/adr-01-02-build-system.md) | necessary — Zig requires its own build system |
| ZMQ binding | `zmq-sys` 0.11.0 (Rust FFI crate wrapping libzmq C API via `pkg-config`) | Zig `@cImport` of libzmq headers (direct C interop) | [ADR-01-03](adr/01-platform/adr-01-03-zmq-binding.md) | necessary — core messaging; eliminates FFI wrapper layer |
| ZMQ context/socket abstraction | Custom `Context`/`Socket` structs in `socket.rs` | Eliminated — absorbed by direct C calls with `defer` cleanup | [ADR-01-03](adr/01-platform/adr-01-03-zmq-binding.md) | eliminated — topology decision; wrappers add indirection without safety benefit in Zig |
| ZMQ message abstraction | Custom `Message` struct in `message.rs` | Eliminated — absorbed by inline `var msg` + `defer zmq_msg_close` | [ADR-01-03](adr/01-platform/adr-01-03-zmq-binding.md) | eliminated — topology decision; RAII via `defer` replaces wrapper struct |
| ZMQ error handling | Custom `Error` enum in `error.rs` (27 variants) | Eliminated — absorbed by inline `zmq_errno()` + `zmq_strerror()` at error sites | [ADR-01-03](adr/01-platform/adr-01-03-zmq-binding.md) | eliminated — topology decision; Zig calls C error functions directly |
| Relay loop | Blocking `zmq_msg_recv` → `zmq_msg_send` in spawned OS thread | Blocking recv → send in dedicated Zig `std.Thread` | [ADR-02-01](adr/02-runtime/adr-02-01-relay-loop.md) | necessary — core relay logic; performance-critical hot path |
| Metrics client | `statsd` 0.16.0 crate (UDP StatsD) | Hand-rolled Zig StatsD UDP client (~30 lines) | [ADR-02-04](adr/02-runtime/adr-02-04-observability.md) | necessary — operational contract requires StatsD metrics |
| Memory monitoring | `procfs` 0.15.1 crate (reads `/proc/self/stat` RSS) | Direct `/proc/self/statm` read via Zig `std.fs` (~10 lines) | [ADR-02-04](adr/02-runtime/adr-02-04-observability.md) | necessary — metrics contract; eliminates heavy dependency tree |
| Logging implementation | Custom `log::Log` trait impl using `colored` crate | Custom `std.log` function with inline ANSI escape codes | [ADR-02-04](adr/02-runtime/adr-02-04-observability.md) | necessary — operational logging with contracted format |
| Coloured terminal output | `colored` 1.6.1 crate | Eliminated — absorbed by inline ANSI escape sequences in log function | [ADR-02-04](adr/02-runtime/adr-02-04-observability.md) | artifact — absorbed into logging; no separate component |
| POSIX signal handling | `libc` 0.2.154 crate (`sigwait`, `signal`, `raise`) | Zig `std.c` / `std.os.linux` signal functions | [ADR-02-02](adr/02-runtime/adr-02-02-lifecycle-shutdown.md) | necessary — daemon lifecycle requires SIGTERM handling |
| Cross-thread coordination | `Arc<AtomicBool>` / `Arc<AtomicUsize>` | Stack-allocated `std.atomic.Value` with pointer sharing | [ADR-02-03](adr/02-runtime/adr-02-03-threading-coordination.md) | necessary — eliminates Arc heap allocation and refcount overhead |
| Thread management | `std::thread::spawn` + `JoinHandle` in `Option` | `std.Thread.spawn` + `thread.join()` in explicit shutdown | [ADR-02-03](adr/02-runtime/adr-02-03-threading-coordination.md) | necessary — relay and metrics require dedicated OS threads |
| systemd sd_notify | Unix datagram via `std::os::unix::net::UnixDatagram` | Zig `std.posix.sendto` on Unix socket | [ADR-02-02](adr/02-runtime/adr-02-02-lifecycle-shutdown.md) | necessary — systemd integration is a hard contract (invariant 3) |
| Configuration loading | `std::env::var_os` with fallback helpers | `std.posix.getenv` with fallback helpers | [ADR-02-02](adr/02-runtime/adr-02-02-lifecycle-shutdown.md) | necessary — runtime config; trivial in any language |
| Debian packaging | `.deb` via `dpkg-deb` with systemd units | Same `.deb` structure; build source path updated | [ADR-03-01](adr/03-packaging/adr-03-01-debian-docker-packaging.md) | necessary — deployment contract; structure stays identical |
| Docker packaging | Debian sid-slim Dockerfiles installing `.deb` | Same Dockerfile structure; unchanged | [ADR-03-01](adr/03-packaging/adr-03-01-debian-docker-packaging.md) | necessary — deployment contract |
| CI/CD pipeline | CircleCI + GitHub Actions + Jenkins with Rust image | Adapt to Zig compiler image; same pipeline structure | [ADR-03-02](adr/03-packaging/adr-03-02-ci-cd-pipeline.md) | necessary — build pipeline must produce binaries |
| Cross-compilation | Cargo targets with clang-13 LTO | Zig native cross-compilation (`-target`) — no external linker | [ADR-01-02](adr/01-platform/adr-01-02-build-system.md) | necessary — multi-arch is a hard requirement |
| Blackbox tests | Python behave BDD (6 features, ~10 scenarios) | Unchanged — language-agnostic | [ADR-03-01](adr/03-packaging/adr-03-01-debian-docker-packaging.md) | necessary — tests must pass unchanged (invariant 6) |
| Performance tests | Python perf suite (1K–1B messages) | Unchanged — language-agnostic | [ADR-02-01](adr/02-runtime/adr-02-01-relay-loop.md) | necessary — performance regression gate (invariant 7) |
| libzmq runtime dependency | Dynamic link to `libzmq5 >= 4.3, < 4.4` | Dynamic link preserved | [ADR-01-04](adr/01-platform/adr-01-04-libzmq-linking.md) | necessary — ZMQ C library is the messaging substrate |
| Rust formatter config | `.rustfmt.toml` (edition 2018) | Removed | [ADR-01-01](adr/01-platform/adr-01-01-language-zig.md) | artifact — Rust-specific; removed during conversion |
| Clippy lint config | `#[allow(clippy::*)]` attributes in source | Removed | [ADR-01-01](adr/01-platform/adr-01-01-language-zig.md) | artifact — Rust-specific; removed during conversion |
| Dependabot cargo config | `.github/dependabot.yml` for cargo ecosystem | Removed | [ADR-03-02](adr/03-packaging/adr-03-02-ci-cd-pipeline.md) | artifact — Cargo-specific; no Zig package registry dependencies |
