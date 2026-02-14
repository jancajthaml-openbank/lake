# Prima Materia — Lake

Lake is a Rust-based ZMQ message relay for the OpenBank platform. It receives messages on a ZMTP 3.0 PULL socket and fans them out via a PUB socket — a stateless, high-throughput, single-binary service. The architectural intent is to convert the Rust implementation to hand-written arm64 (AArch64) assembly, eliminating the Rust compiler and standard library from the runtime while preserving identical external behaviour, wire protocol compatibility, and operational contracts.

## Domain capabilities

| Domain | Purpose | Cloud touchpoints |
|--------|---------|-------------------|
| Message relay | Receives messages from upstream services and fans out to downstream subscribers over ZMTP 3.0 TCP | None |
| Configuration | Reads runtime parameters (ports, log level, metrics endpoint) from environment variables at startup | None |
| Process lifecycle | Manages startup, systemd readiness notification, and SIGTERM-based graceful shutdown | None |
| Metrics emission | Reports message throughput count and memory usage to a StatsD daemon over UDP every second | None |
| Logging | Writes timestamped, colour-coded log messages to stdout | None |
| Error handling | Maps ZMQ/POSIX errno codes to human-readable messages for log output | None |
| Packaging & deployment | Produces Debian .deb packages and Docker images for amd64 and arm64 targets | None (Docker Hub used at CI publish time only) |
| CI/CD | Multi-arch build, test, package, and publish pipelines across CircleCI, GitHub Actions, and Jenkins | None (all CI-time only) |
| Black-box testing | BDD integration tests verifying install, configuration, systemd management, metrics, and message relay | None |
| Performance testing | Throughput benchmarking from 1K to 1B messages measuring relay rate | None |

## Current → desired mapping

| Capability | Current | Desired | ADR | Challenged |
|-----------|---------|---------|-----|-----------|
| Core relay loop (PULL→PUB) | Rust unsafe FFI to libzmq (`zmq_msg_recv`/`zmq_msg_send` via `zmq-sys` crate) | arm64 assembly with direct C ABI calls to libzmq | TBD | TBD |
| ZMQ context management | Rust `Context` struct wrapping `zmq_ctx_new`/`zmq_ctx_set`/`zmq_ctx_term` via `zmq-sys` | arm64 assembly with direct `BL` calls to libzmq functions | TBD | TBD |
| ZMQ socket management | Rust `Socket` struct wrapping `zmq_socket`/`zmq_bind`/`zmq_connect`/`zmq_setsockopt`/`zmq_close` via `zmq-sys` | arm64 assembly with direct `BL` calls to libzmq functions | TBD | TBD |
| ZMQ message handling | Rust `Message` struct with `zmq_msg_init`/`zmq_msg_close` via `zmq-sys`, `Drop` for cleanup | arm64 assembly with manual `zmq_msg_t` struct management on stack | TBD | TBD |
| Error code mapping | Rust `Error` enum mapping `zmq_errno()` to variants with `zmq_strerror()` | arm64 assembly with register-based errno checks and `zmq_strerror()` calls | TBD | TBD |
| Environment variable configuration | Rust `std::env::var_os` + string parsing (`config.rs`) | arm64 assembly calling libc `getenv` + manual integer parsing | TBD | TBD |
| Logging to stdout | Rust custom `Logger` implementing `log::Log` trait with `colored` crate, manual date calculation (`logger.rs`) | arm64 assembly writing formatted ANSI-escaped strings via `write` syscall to fd 1 | TBD | TBD |
| StatsD metrics emission | Rust `statsd` crate (UDP socket, pipeline batching) (`metrics.rs`) | arm64 assembly with raw UDP `socket`/`sendto` syscalls, manual StatsD protocol formatting | TBD | TBD |
| Memory usage measurement | Rust `procfs` crate reading `/proc/self/stat` RSS (`metrics.rs`, linux-only) | arm64 assembly reading and parsing `/proc/self/stat` via `open`/`read`/`close` syscalls | TBD | TBD |
| Systemd notification | Rust `UnixDatagram::send_to` to `$NOTIFY_SOCKET` (`program.rs`) | arm64 assembly with `socket`/`sendto` syscalls to Unix datagram | TBD | TBD |
| Signal handling (SIGTERM) | Rust `libc::signal`/`libc::sigwait`/`libc::sigfillset`/`libc::sigaddset` (`main.rs`) | arm64 assembly with `rt_sigaction`/`rt_sigprocmask`/`rt_sigtimedwait` syscalls | TBD | TBD |
| Thread management | Rust `std::thread::spawn` (2 threads: relay, metrics) | arm64 assembly using `clone` syscall or libc `pthread_create` | TBD | TBD |
| Cross-thread atomic coordination | Rust `Arc<AtomicBool>` / `Arc<AtomicUsize>` with `Ordering::Relaxed` | arm64 assembly using `LDAR`/`STLR` (acquire/release) or `LDXR`/`STXR` (exclusive) instructions | TBD | TBD |
| Graceful shutdown (poison pill) | Rust ZMQ PUSH to loopback PULL port to unblock recv (`relay.rs`) | arm64 assembly with same libzmq `zmq_socket`/`zmq_connect`/`zmq_msg_send` C ABI calls | TBD | TBD |
| Build system | Cargo + rustc + clang-13/LLD (LTO, release opt-level 3) | GNU `as` assembler + `ld` linker (or `gcc`/`clang` as assembler driver), Makefile | TBD | TBD |
| Binary stripping | `objcopy --strip-unneeded` / `aarch64-linux-gnu-objcopy --strip-unneeded` | `aarch64-linux-gnu-strip` or `objcopy --strip-all` on assembled binary | TBD | TBD |
| Debian packaging | `dpkg-buildpackage` with Rust-compiled binary, `lake.install.arm64` | `dpkg-buildpackage` with assembled binary (packaging largely unchanged) | TBD | TBD |
| Docker image | `arm64v8/debian:sid-slim` + .deb install | `arm64v8/debian:sid-slim` + .deb install (unchanged) | TBD | TBD |
| Systemd unit files | `lake.service`, `lake-relay.service`, `lake-watcher.path`, `lake-watcher.service` | Unchanged — assembly binary is a drop-in replacement | TBD | TBD |
| Runtime library dependency | `libzmq5 (>= 4.3~, << 4.4~)`, `libc6` | `libzmq5 (>= 4.3~, << 4.4~)`, `libc6` (unchanged — assembly still links dynamically) | TBD | TBD |
| CI/CD pipelines | CircleCI + GitHub Actions + Jenkins with Cargo-based build stages | Pipelines updated to use assembler instead of Cargo; test stages unchanged | TBD | TBD |
| Black-box test suite | Python/Behave BDD tests exercising .deb package via systemd | Unchanged — tests exercise the binary as a black box | TBD | TBD |
| Performance test suite | Python harness with multiprocessing ZMQ workers | Unchanged — tests exercise the binary externally | TBD | TBD |
