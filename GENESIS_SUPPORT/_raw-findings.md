# Phase 01: Discovery — Raw Findings

## 1. Directory Survey

**Project**: `lake` — a ZMQ-backed message relay for the OpenBank platform.

**Language**: Rust (edition 2021, package name `main`, version 0.2.0)
**Build system**: Cargo (Cargo.toml + Cargo.lock)
**Package manager**: Cargo (crates.io registry)

**Top-level structure**:

```
/workspace
├── services/lake/          # Rust source (8 .rs files)
│   ├── Cargo.toml
│   ├── Cargo.lock
│   └── src/
│       ├── main.rs         # Entry point, signal handling
│       ├── config.rs       # Env-based configuration
│       ├── program.rs      # Lifecycle, systemd notify
│       ├── relay.rs        # Core ZMQ PULL→PUB relay loop
│       ├── socket.rs       # ZMQ context/socket wrappers
│       ├── message.rs      # ZMQ message wrapper
│       ├── error.rs        # ZMQ errno mapping
│       └── logger.rs       # Custom log implementation
├── packaging/              # Debian & Docker packaging
│   ├── debian/             # .deb control, systemd units, rules
│   ├── docker/             # amd64 + arm64 Dockerfiles
│   └── init.conf           # Default env config
├── bbtest/                 # Black-box tests (Python/Behave)
├── perf/                   # Performance tests (Python/ZMQ)
├── dev/lifecycle/          # Build/test/lint/release scripts (bash)
├── .circleci/config.yml    # CircleCI pipeline
├── .github/workflows/      # GitHub Actions (health, DevSkim)
├── .jenkins/               # Jenkins pipeline (Groovy)
├── Makefile                # Top-level orchestration
├── docker-compose.yml      # Dev/CI service definitions
└── README.md
```

**Infrastructure artifacts**: Dockerfiles (amd64, arm64), docker-compose.yml, systemd unit files (lake.service, lake-relay.service, lake-watcher.path, lake-watcher.service), Debian packaging (control, rules, install files), CircleCI config, GitHub Actions workflows, Jenkins Groovyscript.

---

## 2. Dependency Scan

### Direct Rust dependencies (Cargo.toml)

| Crate | Version | Classification |
|-------|---------|---------------|
| `zmq-sys` | 0.11.0 | **Messaging** — raw FFI bindings to libzmq (C library) |
| `libc` | 0.2.154 | **System** — POSIX/libc FFI (signals, socket types) |
| `statsd` | 0.16.0 | **Metrics** — StatsD UDP client |
| `log` | 0.4.18 | **General** — logging facade |
| `colored` | 1.6.1 | **General** — terminal color output |
| `procfs` | 0.15.1 | **System** — Linux /proc filesystem reader (linux-only, conditional) |

### Key transitive dependencies (from Cargo.lock)

| Crate | Role |
|-------|------|
| `metadeps` / `pkg-config` | Build-time: locates system libzmq via pkg-config |
| `chrono` | Used by `procfs` for timestamps |
| `rand` | Used by `statsd` for sampling |
| `cc` / `cxx` / `cxx-build` | Build-time: C/C++ compilation (transitive from procfs/chrono) |

### System library dependency

| Library | Version constraint | Source |
|---------|-------------------|--------|
| `libzmq5` | >= 4.3, < 4.4 | `packaging/debian/control` Depends field |

**Classification summary**: No cloud SDKs. No database drivers. No auth libraries. All dependencies are system-level or messaging-related. The project links dynamically against `libzmq5` (a C library).

---

## 3. Cloud Touchpoint Grep

**Result**: No cloud vendor SDKs, no cloud-specific environment variables, no cloud service references found anywhere in the Rust source, configuration, or packaging. The project is already vendor-independent.

The only external service reference is:
- **Docker Hub** (`docker.io/openbank/lake`) — for image publishing (CI/CD only, not runtime)
- **GitHub Packages** (`docker.pkg.github.com`) — for image publishing (CI/CD only, not runtime)
- **GitHub API** — for release asset uploading (CI/CD only)
- **Artifactory** — referenced in Jenkins pipeline for artifact storage (CI/CD only)

**Runtime cloud touchpoints**: None.

---

## 4. Communication Patterns

### Core relay pattern

The system implements a single communication pattern: **ZMQ PULL→PUB relay**.

- **PULL socket** binds to `tcp://0.0.0.0:{LAKE_PORT_PULL}` (default 5562). Receives messages from upstream services (ZMQ PUSH clients).
- **PUB socket** binds to `tcp://0.0.0.0:{LAKE_PORT_PUB}` (default 5561). Publishes messages to downstream subscribers (ZMQ SUB clients).
- **Protocol**: ZMTP 3.0 with NULL security mechanism (no encryption, no auth).
- **Message format**: Opaque binary blobs — the relay does not inspect or transform message content. Topic-based subscription is used by subscribers (first space-delimited token is the topic).

### Socket options

| Socket | Option | Value | Purpose |
|--------|--------|-------|---------|
| PULL | ZMQ_CONFLATE | 0 | Keep all messages (no conflation) |
| PULL | ZMQ_IMMEDIATE | 1 | Only queue to completed connections |
| PULL | ZMQ_LINGER | 0 | Discard pending on close |
| PULL | ZMQ_RCVHWM | 0 | Unlimited receive high-water mark |
| PUB | ZMQ_CONFLATE | 0 | Keep all messages |
| PUB | ZMQ_IMMEDIATE | 1 | Only queue to completed connections |
| PUB | ZMQ_LINGER | 0 | Discard pending on close |
| PUB | ZMQ_SNDHWM | 0 | Unlimited send high-water mark |
| PUB | ZMQ_XPUB_NODROP | 1 | Do not silently drop messages |

### Graceful shutdown

A poison-pill mechanism: on shutdown, a ZMQ PUSH socket connects to the PULL port and sends an empty message to unblock the blocking `zmq_msg_recv` in the relay loop. After the relay thread terminates, `SIGTERM` is raised.

### Systemd notification

The program sends `READY=1` and `STOPPING=1` to the systemd `NOTIFY_SOCKET` via Unix datagram, enabling `Type=notify` service management.

---

## 5. Storage Patterns

**Result**: No databases, no object storage, no caching layers, no session stores. The system is purely a stateless message relay. No data is persisted.

---

## 6. Authentication & Authorization

**Result**: ZMQ NULL mechanism — no authentication or encryption. No identity providers, no token handling, no RBAC/ABAC. The relay accepts all connections on the bound ports.

---

## 7. Configuration & Secrets

### Environment variables

| Variable | Default | Purpose | Source |
|----------|---------|---------|--------|
| `LAKE_PORT_PULL` | 5562 | ZMQ PULL socket bind port | `config.rs`, `init.conf` |
| `LAKE_PORT_PUB` | 5561 | ZMQ PUB socket bind port | `config.rs`, `init.conf` |
| `LAKE_LOG_LEVEL` | INFO | Log verbosity (DEBUG/INFO/WARN/ERROR) | `config.rs`, `init.conf` |
| `LAKE_STATSD_ENDPOINT` | 127.0.0.1:8125 | StatsD daemon address | `config.rs`, `init.conf` |
| `NOTIFY_SOCKET` | (system) | Systemd notification socket | `program.rs` |

### Config file

`/etc/lake/conf.d/init.conf` — sourced by systemd as `EnvironmentFile`. Contains key=value pairs for all `LAKE_*` variables. A systemd `.path` unit watches this directory for changes and triggers restart.

### Secrets

No secrets managed. No secret manager integration. No TLS certificates.

---

## 8. Infrastructure

### Systemd units

| Unit | Type | Purpose |
|------|------|---------|
| `lake.service` | oneshot | Control group parent |
| `lake-relay.service` | notify | Core relay process (`/usr/bin/lake`) |
| `lake-watcher.path` | path | Watches `/etc/lake/conf.d` for changes |
| `lake-watcher.service` | simple | Restarts `lake.service` on config change |

Key service properties: `Restart=always`, `RestartSec=0`, `LimitNOFILE=1048576`, `LimitNPROC=infinity`, `KillSignal=SIGTERM`, `TimeoutStopSec=3`.

### Debian packaging

- Package name: `lake`
- Architecture: `any` (built for amd64 and arm64)
- Runtime dependency: `libzmq5 (>= 4.3~, << 4.4~)`, `libc6`
- Binary installed to `/usr/bin/lake` (renamed from `lake-linux-{arch}`)
- Config installed to `/etc/lake/conf.d/init.conf`

### Docker images

- `amd64`: Based on `amd64/debian:sid-slim`, installs the .deb package
- `arm64`: Based on `arm64v8/debian:sid-slim`, installs the .deb package
- Both use `ENTRYPOINT ["lake"]`

### Build toolchain

- Rust compiler via custom Docker image (`jancajthaml/rust:{arch}`)
- Cross-compilation: clang-13 linker, LLD, LTO enabled
- Strip: `objcopy --strip-unneeded` (amd64), `aarch64-linux-gnu-objcopy --strip-unneeded` (arm64)
- Release profile: `opt-level = 3`

### CI/CD pipelines

| Platform | Purpose | Architectures |
|----------|---------|---------------|
| CircleCI | Primary CI: unit test, compile, debian package, docker package, bbtest, publish, release | amd64 + arm64 |
| GitHub Actions (health.yml) | From-scratch build + unit test on ubuntu-latest | amd64 only |
| GitHub Actions (dev-skim.yml) | Security scanning (DevSkim) | N/A |
| Jenkins | Alternative CI with Artifactory publishing | amd64 (primarily) |

### Multi-arch build

The project already builds for both amd64 and arm64. The Makefile auto-detects architecture via `uname -m`. CircleCI uses `machine-arm64` and `machine-amd64` executors for native builds.

---

## 9. Multi-tenancy Signals

**Result**: No multi-tenancy. No tenant models, no database switching, no subdomain routing. The relay is a single-purpose, single-instance service.

---

## 10. Language and Framework Patterns

### Rust-specific patterns

| Pattern | Type | Where used | Frequency | Structural |
|---------|------|-----------|-----------|-----------|
| Unsafe FFI calls to zmq-sys | FFI/unsafe | `relay.rs`, `socket.rs`, `message.rs`, `error.rs`, `main.rs` | 31 instances across 6 files | Yes — the entire relay loop and all socket operations are unsafe FFI |
| Raw pointer management (`*mut c_void`) | Unsafe memory | `socket.rs` (Context, Socket structs) | Pervasive in socket.rs | Yes — ZMQ handles are raw C pointers |
| Manual `Drop` implementations | RAII | `socket.rs`, `relay.rs`, `message.rs`, `program.rs` | 5 Drop impls | Yes — resource cleanup for ZMQ contexts, sockets, messages, relay thread |
| `unsafe impl Send + Sync` for Context | Concurrency marker | `socket.rs:10-11` | 1 instance | Yes — enables cross-thread ZMQ context sharing |
| `MaybeUninit` for signal handling | Unsafe init | `main.rs:41-54` | 2 instances | Yes — signal set and signum initialization |
| `extern "C" fn` signal handler | FFI callback | `main.rs:58` | 1 instance | Yes — SIGTERM handler |
| `libc::signal`, `libc::sigwait`, `libc::sigfillset`, `libc::sigaddset` | POSIX signal API | `main.rs:34-55` | 4 calls | Yes — signal-based lifecycle management |
| `libc::raise(SIGTERM)` for self-termination | POSIX signal | `relay.rs:97`, `metrics.rs:64` | 2 instances | Yes — thread-to-main-thread communication |
| Conditional compilation (`#[cfg(target_os)]`) | Platform dispatch | `metrics.rs:79-95`, `program.rs:37-57` | 2 modules | Medium — platform-specific memory reading and systemd notify |
| Custom logger with manual timestamp formatting | No-framework logging | `logger.rs` (entire file) | 1 implementation | Yes — avoids chrono dependency at runtime, manual date calculation |
| `Arc<AtomicBool>` for cross-thread cancellation | Atomic coordination | `main.rs`, `relay.rs`, `metrics.rs`, `program.rs` | Pervasive | Yes — the shutdown coordination mechanism |
| `Arc<AtomicUsize>` for lock-free counter | Atomic metrics | `metrics.rs:12` | 1 instance | Medium — metrics accumulation |
| `thread::spawn` for concurrency | OS threads | `relay.rs:53`, `metrics.rs:38` | 2 threads | Yes — relay loop and metrics reporting run in separate OS threads |
| `zmq_sys::zmq_msg_t::default()` for zero-init | FFI struct init | `message.rs:20` | 1 instance | Medium — relies on Default trait for C struct |
| `mem::transmute` for lifetime extension | Unsafe cast | `error.rs:123` | 1 instance | Low — casts CStr bytes to static lifetime |

### External library conventions

| Convention | Where | Notes |
|-----------|-------|-------|
| `statsd::Client` UDP pipeline | `metrics.rs` | Fire-and-forget UDP metrics every 1 second |
| `log` facade macros | All source files | `log::info!`, `log::debug!`, `log::error!`, `log::warn!` |
| `colored` for terminal output | `logger.rs` | Color-coded log levels |
| `procfs::process::Process::myself()` | `metrics.rs` | Linux-only memory measurement via /proc |

---

## 11. External Contract Inventory

### ZMQ wire protocol (ZMTP 3.0)

| Contract | Type | Endpoint / schema | Where defined |
|----------|------|------------------|--------------|
| PULL socket (message ingress) | ZMQ/ZMTP 3.0 PULL | `tcp://0.0.0.0:5562` | `relay.rs:128`, `config.rs:19` |
| PUB socket (message egress) | ZMQ/ZMTP 3.0 PUB | `tcp://0.0.0.0:5561` | `relay.rs:162`, `config.rs:20` |
| Shutdown poison pill | ZMQ/ZMTP 3.0 PUSH→PULL | `tcp://127.0.0.1:{pull_port}` | `relay.rs:24` |
| StatsD metrics | UDP/StatsD | `127.0.0.1:8125` | `metrics.rs:39`, `config.rs:22` |
| Systemd notification | Unix datagram | `$NOTIFY_SOCKET` | `program.rs:38-52` |

### StatsD metrics

| Metric | Type | Description |
|--------|------|-------------|
| `openbank.lake.message.relayed` | count | Messages relayed per second |
| `openbank.lake.memory.bytes` | gauge | RSS memory in bytes |

### CLI interface

| Command | Arguments | Description |
|---------|-----------|-------------|
| `lake` | (none — configured via env) | Starts the relay daemon |

---

## 12. Test Architecture

| Aspect | Finding |
|--------|---------|
| **Unit test framework** | Cargo's built-in `#[test]` (invoked via `cargo test`). No unit tests found in the source files — the codebase has zero Rust test functions. |
| **Black-box test framework** | Python 3 + Behave (BDD/Gherkin). 6 feature files, ~15 scenarios. |
| **Black-box test runner** | `bbtest/main.py` — orchestrates Behave, converts results to Cucumber JSON, then to JUnit XML. |
| **Black-box test helpers** | `openbank_testkit` (Shell, Package, Platform, StatsdMock), `systemd.journal` (journalctl reader), custom `ZMQHelper` (push/subscribe client), `eventually` (polling retry decorator). |
| **Black-box test categories** | Install/uninstall (.deb), systemd unit management, configuration, metrics collection, message relay order. |
| **Performance test framework** | Python 3 custom harness (`perf/main.py`). Pushes increasing message volumes (1K to 1B) and measures throughput. |
| **Performance test helpers** | `Publisher` (multiprocessing ZMQ push+subscribe workers), `ApplianceManager` (systemd-based service lifecycle), `Metrics`/`Graph` (JSON metrics + matplotlib plotting). |
| **Test execution environment** | Docker container (`jancajthaml/bbtest:{arch}`), runs systemd inside container, installs .deb, exercises real systemd service lifecycle. |
| **Test distribution** | 0 unit tests, ~15 black-box integration tests, 1 performance test suite. |

---

## 13. Module Dependency Mapping

The Rust source is a single flat module structure (no sub-crates, no workspace):

```
main.rs (entry point)
├── config.rs     (leaf — no internal deps)
├── error.rs      (leaf — depends only on zmq_sys)
├── logger.rs     (leaf — depends only on colored, log)
├── message.rs    (depends on: error)
├── metrics.rs    (depends on: config)
├── program.rs    (depends on: config, logger)
├── socket.rs     (depends on: error)
└── relay.rs      (depends on: config, error, message, metrics, socket)
```

**Dependency graph**:

```
config ←── program
  │           └── logger
  │
  ├── metrics
  │
  └── relay ──→ message ──→ error
       │                     ↑
       └──→ socket ──────────┘
```

**Leaf modules** (zero internal dependants): `logger`, `error`, `config`
**Core modules** (most dependants): `relay` (depends on 5 other modules)
**Circular dependencies**: None

### External module dependencies

| Rust module | External crate dependencies |
|------------|---------------------------|
| `config.rs` | `std::env` |
| `error.rs` | `zmq_sys::errno`, `zmq_sys::zmq_strerror`, `std::ffi`, `std::fmt`, `std::mem`, `std::str` |
| `logger.rs` | `colored::Colorize`, `log::{Level, LevelFilter, Log, Metadata, Record, SetLoggerError}`, `std::time::SystemTime` |
| `message.rs` | `zmq_sys::zmq_msg_t`, `zmq_sys::zmq_msg_init`, `zmq_sys::zmq_msg_close` |
| `metrics.rs` | `statsd::Client`, `std::sync::atomic`, `std::thread`, `std::time::Duration`, `procfs` (linux) |
| `program.rs` | `std::os::unix::net::UnixDatagram`, `std::sync::atomic`, `std::env`, `std::io`, `log` |
| `relay.rs` | `zmq_sys::zmq_msg_recv`, `zmq_sys::zmq_msg_send`, `zmq_sys::ZMQ_PULL`, `zmq_sys::ZMQ_PUB`, `zmq_sys::ZMQ_PUSH`, `std::thread`, `std::sync::atomic`, `libc::raise`, `libc::SIGTERM` |
| `socket.rs` | `zmq_sys::zmq_ctx_new`, `zmq_sys::zmq_ctx_set`, `zmq_sys::zmq_ctx_term`, `zmq_sys::zmq_socket`, `zmq_sys::zmq_bind`, `zmq_sys::zmq_connect`, `zmq_sys::zmq_setsockopt`, `zmq_sys::zmq_close`, `zmq_sys::zmq_errno`, `libc::c_void`, `libc::size_t`, `std::ffi`, `std::mem` |
| `main.rs` | `std::mem`, `std::sync::atomic::Ordering`, `libc::signal`, `libc::sigwait`, `libc::sigfillset`, `libc::sigaddset`, `libc::SIGTERM`, `libc::sigset_t`, `libc::sighandler_t`, `libc::c_int`, `libc::c_void` |

---

## Domain Capabilities Table

| Domain | Key files / classes | Purpose | Cloud touchpoints |
|--------|-------------------|---------|-------------------|
| Message relay | `relay.rs`, `socket.rs`, `message.rs` | Receives messages on a PULL socket and fans out via PUB socket over ZMTP 3.0 TCP | None |
| Configuration | `config.rs`, `packaging/init.conf` | Environment-variable-based configuration for ports, log level, and metrics endpoint | None |
| Lifecycle management | `program.rs`, `main.rs` | Process startup, systemd notification, SIGTERM-based graceful shutdown | None |
| Metrics | `metrics.rs` | Emits message relay count and memory usage to a StatsD endpoint every second | None |
| Logging | `logger.rs` | Custom colored log output with manual timestamp formatting to stdout | None |
| Error handling | `error.rs` | Maps ZMQ errno codes to Rust enum variants with human-readable messages | None |
| Packaging & deployment | `packaging/`, `Makefile`, `docker-compose.yml` | Debian .deb packaging, Docker images, systemd service units for amd64 and arm64 | Docker Hub (CI publish only) |
| CI/CD | `.circleci/`, `.github/workflows/`, `.jenkins/` | Multi-arch build, test, package, publish pipelines | GitHub Actions, CircleCI, Docker Hub, GitHub Packages (CI only) |
| Black-box testing | `bbtest/` | BDD tests for install, config, management, metrics, and relay behaviour | None |
| Performance testing | `perf/` | Throughput benchmarking up to 1B messages | None |

## Communication Patterns Table

| Pattern | Protocol | Where used | Cloud dependency |
|---------|----------|-----------|-----------------|
| ZMQ PULL (message ingress) | ZMTP 3.0 over TCP | `relay.rs:107-134` | None |
| ZMQ PUB (message egress) | ZMTP 3.0 over TCP | `relay.rs:137-169` | None |
| ZMQ PUSH (shutdown signal) | ZMTP 3.0 over TCP (loopback) | `relay.rs:21-31` | None |
| StatsD metrics emission | UDP | `metrics.rs:38-65` | None |
| Systemd notify | Unix datagram | `program.rs:38-52` | None |
| Signal handling (SIGTERM) | POSIX signals | `main.rs:34-58` | None |

## Language Patterns Table

| Pattern | Type | Where used | Frequency | Structural |
|---------|------|-----------|-----------|-----------|
| Unsafe FFI calls to C library (libzmq) | FFI/unsafe | `relay.rs`, `socket.rs`, `message.rs`, `error.rs`, `main.rs` | 31 instances / 6 files | Yes |
| Raw C pointer management (`*mut c_void`) | Unsafe memory | `socket.rs` (Context.underlying, Socket.sock) | 2 structs, pervasive use | Yes |
| Manual `Drop` for RAII cleanup | Resource management | `socket.rs`, `relay.rs`, `message.rs`, `program.rs` | 5 implementations | Yes |
| `unsafe impl Send + Sync` | Concurrency safety | `socket.rs:10-11` | 1 instance | Yes |
| `MaybeUninit` + `assume_init` | Unsafe initialization | `main.rs:41-54` | 2 instances | Yes |
| `extern "C" fn` callback | FFI callback | `main.rs:58` | 1 instance | Yes |
| POSIX signal API via libc | System interface | `main.rs:34-55` | 4 libc calls | Yes |
| `libc::raise(SIGTERM)` for thread self-kill | System interface | `relay.rs:97`, `metrics.rs:64` | 2 instances | Yes |
| Conditional compilation (`#[cfg(...)]`) | Platform dispatch | `metrics.rs`, `program.rs` | 4 cfg blocks | Medium |
| Manual date/time formatting (no chrono at runtime) | Custom implementation | `logger.rs:55-100` | 1 implementation | Medium |
| `Arc<AtomicBool>` cross-thread cancellation | Lock-free coordination | `main.rs`, `relay.rs`, `metrics.rs`, `program.rs` | Pervasive | Yes |
| `Arc<AtomicUsize>` lock-free counter | Atomic metrics | `metrics.rs:12` | 1 instance | Medium |
| `thread::spawn` OS-level concurrency | Threading | `relay.rs:53`, `metrics.rs:38` | 2 threads | Yes |
| `mem::transmute` lifetime extension | Unsafe cast | `error.rs:123` | 1 instance | Low |
| `ffi::CString` for C string interop | FFI | `socket.rs:63,71` | 2 instances | Medium |

## External Contracts Table

| Contract | Type | Endpoint / schema | Where defined |
|----------|------|------------------|--------------|
| ZMQ PULL ingress | ZMTP 3.0 / TCP | `tcp://0.0.0.0:{LAKE_PORT_PULL}` (default 5562) | `relay.rs:128`, `config.rs:19` |
| ZMQ PUB egress | ZMTP 3.0 / TCP | `tcp://0.0.0.0:{LAKE_PORT_PUB}` (default 5561) | `relay.rs:162`, `config.rs:20` |
| StatsD metrics | UDP / StatsD protocol | `{LAKE_STATSD_ENDPOINT}` (default 127.0.0.1:8125) | `metrics.rs:39`, `config.rs:22` |
| Systemd notification | Unix datagram | `$NOTIFY_SOCKET` | `program.rs:38-52` |
| `lake` CLI | Binary executable | `/usr/bin/lake` (no arguments, env-configured) | `packaging/debian/lake-relay.service:12` |
| StatsD metric: `openbank.lake.message.relayed` | StatsD count | Per-second relay count | `metrics.rs:57` |
| StatsD metric: `openbank.lake.memory.bytes` | StatsD gauge | RSS memory bytes | `metrics.rs:55` |

## Test Architecture Table

| Aspect | Finding |
|--------|---------|
| Unit test framework | Cargo built-in `#[test]` — **zero unit tests exist** |
| Black-box framework | Python 3 / Behave (Gherkin BDD) |
| Black-box scenarios | ~15 across 6 feature files (install, uninstall, config, management, metrics, relay) |
| Performance framework | Custom Python harness with multiprocessing ZMQ workers |
| Test execution | Docker container with systemd, installs .deb, exercises real service lifecycle |
| Test dependencies | `openbank_testkit`, `behave`, `behave2cucumber`, `zmq` (pyzmq), `systemd.journal` |
| Coverage | No code coverage tooling configured |
| CI test stages | Unit test → blackbox test (blocking), performance test (optional) |

## Module Dependency Graph

```
config (leaf)
├──→ program ──→ logger (leaf)
├──→ metrics
└──→ relay ──→ message ──→ error (leaf)
      └──→ socket ──→ error (leaf)
```

Leaf modules: `config`, `logger`, `error`
Core module: `relay` (5 internal dependencies)
Cycles: None
