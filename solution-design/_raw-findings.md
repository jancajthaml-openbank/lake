# Raw Findings — Lake (Phase 01 Discovery)

Project: **lake** — ZMQ-based message relay for the OpenBank platform.
Repository: `jancajthaml-openbank/lake`

## Directory survey

```
/workspace
├── services/lake/           # Rust source (single binary crate)
│   ├── Cargo.toml
│   ├── Cargo.lock
│   └── src/
│       ├── main.rs          # Entry point, signal handling
│       ├── config.rs         # Env-based configuration
│       ├── error.rs          # ZMQ error codes enum + Display
│       ├── logger.rs         # Custom log::Log impl with colored output
│       ├── message.rs        # ZMQ message wrapper
│       ├── metrics.rs        # StatsD metrics reporter thread
│       ├── program.rs        # Lifecycle (systemd notify, running flag)
│       ├── relay.rs          # Core relay: ZMQ PULL → PUB forwarding
│       └── socket.rs         # Low-level ZMQ context/socket wrappers
├── packaging/
│   ├── docker/amd64/Dockerfile
│   ├── docker/arm64/Dockerfile
│   ├── debian/               # Full .deb packaging (control, rules, systemd units)
│   ├── init.conf             # Default env config
│   └── bin/                  # Build artifact output
├── bbtest/                   # Blackbox tests (Python behave)
├── perf/                     # Performance tests (Python + ZMQ)
├── dev/lifecycle/            # Build lifecycle scripts (bash)
├── .circleci/config.yml      # Primary CI pipeline
├── .github/workflows/        # Health check + DevSkim security scan
├── .jenkins/                 # Jenkins pipeline (legacy/alternative)
├── Makefile                  # Top-level build orchestration
├── docker-compose.yml        # Dev/CI service definitions
├── .rustfmt.toml             # Rust formatter config
└── .editorconfig             # Editor config
```

- **Language**: Rust (edition 2021)
- **Build system**: Cargo (no workspace; single crate at `services/lake/`)
- **Package name**: `main` (binary crate, v0.2.0)
- **Target architectures**: amd64, arm64
- **Deployment**: Debian `.deb` package + Docker images (debian:sid-slim base)

## Dependency scan

| Dependency | Version | Classification | Purpose |
|-----------|---------|---------------|---------|
| `zmq-sys` | 0.11.0 | Messaging (C FFI bindings) | Raw ZMQ C API bindings via `libzmq5` |
| `statsd` | 0.16.0 | Metrics | UDP StatsD client for gauge/counter reporting |
| `log` | 0.4.18 | General (logging facade) | Rust logging facade (features: `std`) |
| `libc` | 0.2.154 | System (FFI) | POSIX signal handling, `sigwait`, `raise` |
| `colored` | 1.6.1 | General (terminal) | Colored terminal output for log levels |
| `procfs` | 0.15.1 | System (Linux-only) | `/proc` reader for RSS memory reporting |

**Runtime system dependency**: `libzmq5 >= 4.3, < 4.4` (declared in `debian/control`).

No cloud SDKs. No database drivers. No HTTP/REST libraries. No auth libraries.

## Domain capabilities table

| Domain | Key files / classes | Purpose | Cloud touchpoints |
|--------|-------------------|---------|-------------------|
| Message relay | `relay.rs`, `socket.rs`, `message.rs` | Receives messages via ZMQ PULL socket and fans them out via ZMQ PUB socket, enabling inter-service communication without direct binding | None |
| Metrics reporting | `metrics.rs` | Periodically reports message throughput (count) and memory usage (RSS bytes) to a StatsD daemon over UDP | None |
| Process lifecycle | `program.rs`, `main.rs` | Manages daemon startup/shutdown with systemd sd_notify integration (READY/STOPPING) and POSIX signal handling (SIGTERM) | None |
| Configuration | `config.rs`, `packaging/init.conf` | Loads runtime parameters from environment variables with sensible defaults | None |
| Logging | `logger.rs` | Custom `log::Log` implementation with ISO-8601 timestamps, colored levels, manual date formatting (no chrono at runtime) | None |
| Error handling | `error.rs` | Maps ZMQ C errno codes to Rust enum with Display for human-readable messages | None |

## Communication patterns table

| Pattern | Protocol | Where used | Cloud dependency |
|---------|----------|-----------|-----------------|
| ZMQ PULL socket (ingest) | ZMTP 3.0 / TCP | `relay.rs:setup_pull_socket` — binds `tcp://0.0.0.0:{LAKE_PORT_PULL}` (default 5562) | None |
| ZMQ PUB socket (fanout) | ZMTP 3.0 / TCP | `relay.rs:setup_pub_socket` — binds `tcp://0.0.0.0:{LAKE_PORT_PUB}` (default 5561) | None |
| ZMQ PUSH (internal kill) | ZMTP 3.0 / TCP | `relay.rs:Drop` — connects to `tcp://127.0.0.1:{pull_port}` to unblock the relay loop on shutdown | None |
| StatsD UDP | UDP | `metrics.rs` — sends gauge/counter to `{LAKE_STATSD_ENDPOINT}` (default `127.0.0.1:8125`) | None |
| systemd sd_notify | Unix datagram | `program.rs:notify` — sends `READY=1` / `STOPPING=1` to `$NOTIFY_SOCKET` | None |

**Security model**: ZMQ NULL mechanism (ZMTP 3.0) — no encryption, no authentication. README states this explicitly.

## Storage patterns

None. The relay is stateless — messages flow through without persistence. No database, no object storage, no caching layer, no session store.

## Authentication & authorization

None. ZMQ uses NULL mechanism (no CURVE, no PLAIN, no GSSAPI). The relay accepts any connection that can reach the TCP port.

## Configuration & secrets

| Variable | Default | Source | Purpose |
|----------|---------|--------|---------|
| `LAKE_PORT_PULL` | 5562 | env / `init.conf` | ZMQ PULL socket bind port |
| `LAKE_PORT_PUB` | 5561 | env / `init.conf` | ZMQ PUB socket bind port |
| `LAKE_LOG_LEVEL` | INFO | env / `init.conf` | Log level (DEBUG/INFO/WARN/ERROR) |
| `LAKE_STATSD_ENDPOINT` | 127.0.0.1:8125 | env / `init.conf` | StatsD daemon UDP endpoint |
| `NOTIFY_SOCKET` | (systemd-provided) | env | systemd notification socket path |

Config is loaded once at startup (`Configuration::load()`). No hot-reload of the binary — the `lake-watcher.path` systemd unit watches `/etc/lake/conf.d` and restarts the service on config change.

No secret manager integration. No sensitive credentials stored.

## Infrastructure

### Systemd units (deployment)

| Unit | Type | Purpose |
|------|------|---------|
| `lake.service` | oneshot (control group) | Parent unit; gates all sub-services on `/etc/lake/conf.d/init.conf` existence |
| `lake-relay.service` | notify | The actual relay binary; `Type=notify`, `Restart=always`, `RestartSec=0`, `LimitNOFILE=1048576` |
| `lake-watcher.path` | path | Watches `/etc/lake/conf.d` for changes |
| `lake-watcher.service` | simple | Restarts `lake.service` when config changes |

### Docker

- Dockerfiles for `amd64` and `arm64` (Debian sid-slim base).
- Installs the `.deb` package inside the container.
- Entrypoint: `lake`.

### CI/CD

- **CircleCI** (`.circleci/config.yml`): Primary pipeline — checkout → unit-test → compile (amd64/arm64) → package-debian → package-docker → blackbox-test → publish → release. Tag-triggered releases. Hourly rolling blackbox contract tests on `main`.
- **GitHub Actions** (`.github/workflows/`): Health check (build from scratch with cargo, installs libzmq from source). DevSkim security scanner.
- **Jenkins** (`.jenkins/commit.groovy`): Alternative pipeline — same stages, publishes to Artifactory.
- **Dependabot**: Daily cargo dependency updates for `services/lake`.

### Build toolchain

- Custom Docker image `jancajthaml/rust:{arch}` for compilation.
- Uses `clang-13` as linker with LTO enabled (`-Clinker-plugin-lto`).
- Cross-compilation: amd64 target `x86_64-unknown-linux-gnu`, arm64 target `aarch64-unknown-linux-gnu`.
- Binary is stripped with `objcopy --strip-unneeded`.
- Release profile: `opt-level = 3`.

## Multi-tenancy signals

None. The relay is a single-tenant infrastructure component. All messages flow through the same PULL→PUB pipeline regardless of origin.

---

## Language patterns table (conversion: Rust → Zig)

| # | Pattern | Type | Where used | Frequency | Structural |
|---|---------|------|-----------|-----------|-----------|
| 1 | `unsafe` blocks for C FFI calls | Unsafe FFI | `socket.rs` (17), `relay.rs` (4), `main.rs` (4), `message.rs` (3), `error.rs` (2), `metrics.rs` (1) | 31 instances across 6 files | Yes — all ZMQ interaction is through unsafe FFI |
| 2 | `unsafe impl Send + Sync` on raw pointer wrapper | Unsafe trait impl | `socket.rs:10-11` (Context struct) | 1 instance | Yes — enables cross-thread sharing of ZMQ context |
| 3 | RAII via `impl Drop` for resource cleanup | Ownership/Drop | `socket.rs` (Context, Socket), `message.rs` (Message), `relay.rs` (Relay), `metrics.rs` (Metrics), `program.rs` (Program) | 6 Drop impls | Yes — core resource management pattern |
| 4 | `Arc<AtomicBool>` for cross-thread running flag | Shared atomic state | `main.rs`, `program.rs`, `relay.rs`, `metrics.rs` | 5 files, pervasive | Yes — coordinates shutdown across threads |
| 5 | `Arc<AtomicUsize>` for lock-free counter | Atomic metrics counter | `metrics.rs` | 1 instance | Yes — message counting without locking |
| 6 | `std::thread::spawn` with `move` closures | OS thread spawning | `relay.rs:53`, `metrics.rs:38` | 2 instances | Yes — relay and metrics each run in dedicated threads |
| 7 | `thread::JoinHandle` stored in `Option` for deferred join | Thread lifecycle | `relay.rs:13`, `metrics.rs:14` | 2 instances | Yes — enables orderly shutdown via Drop |
| 8 | `mem::MaybeUninit` for uninitialized signal sets | Unsafe memory | `main.rs:41,51` | 2 instances | Yes — POSIX signal handling requires uninitialized memory |
| 9 | `mem::transmute` for lifetime extension | Unsafe transmute | `error.rs:123` | 1 instance | No — could use safer alternative |
| 10 | Raw pointer manipulation (`*mut c_void`) | Unsafe pointers | `socket.rs` (Context, Socket fields), `message.rs` (msg_ptr fn) | Pervasive in socket/message layer | Yes — ZMQ C API requires raw pointers |
| 11 | `libc::signal` / `libc::sigwait` for POSIX signals | Unsafe syscalls | `main.rs:34-56` | 1 block | Yes — daemon signal handling |
| 12 | `libc::raise(SIGTERM)` as exit mechanism | Unsafe syscall | `relay.rs:97`, `metrics.rs:64` | 2 instances | Yes — worker threads signal main thread to exit |
| 13 | `#[cfg(target_os)]` conditional compilation | Platform dispatch | `metrics.rs:79,91` (linux/macos mem_bytes), `program.rs:37,54` (unix/non-unix notify) | 4 attributes | Yes — platform-specific behaviour |
| 14 | `log` facade with custom `impl Log` | Trait implementation | `logger.rs` | 1 impl | Yes — all logging goes through this |
| 15 | `std::error::Error` + `Display` + `Debug` trait impls | Trait implementation | `error.rs:129-151` | 3 impls on Error | No — standard Rust pattern |
| 16 | `ffi::CString` for C string conversion | FFI marshalling | `socket.rs:63,71` | 2 instances | Yes — needed for ZMQ C API string parameters |
| 17 | Builder-like config with env fallbacks | Initialization | `config.rs` | 1 struct | No — simple pattern |
| 18 | `#[must_use]` on constructors | Lint attribute | `relay.rs:41`, `metrics.rs:28`, `config.rs:16` | 3 instances | No — documentation hint |
| 19 | `#[allow(clippy::*)]` lint suppressions | Lint attribute | `socket.rs` (3), `metrics.rs` (2), `logger.rs` (1), `error.rs` (1) | 7 instances | No — Clippy noise management |
| 20 | Manual date formatting (no chrono at runtime) | Custom implementation | `logger.rs:60-87` | 1 block | No — could use standard lib |
| 21 | Pointer cast chain (`as *mut c_void as sighandler_t`) | Unsafe cast | `main.rs:37` | 1 instance | Yes — signal handler registration |
| 22 | `zmq_sys::zmq_msg_t::default()` for C struct init | FFI struct init | `message.rs:20` | 1 instance | Yes — ZMQ message allocation |

## External contracts table (conversion)

| # | Contract | Type | Endpoint / schema | Where defined |
|---|----------|------|------------------|--------------|
| 1 | ZMQ PULL ingest | ZMTP 3.0 / TCP | `tcp://0.0.0.0:5562` (configurable via `LAKE_PORT_PULL`) | `relay.rs:128`, `config.rs:19` |
| 2 | ZMQ PUB fanout | ZMTP 3.0 / TCP | `tcp://0.0.0.0:5561` (configurable via `LAKE_PORT_PUB`) | `relay.rs:162`, `config.rs:20` |
| 3 | StatsD UDP metrics | UDP (StatsD protocol) | `{LAKE_STATSD_ENDPOINT}` (default `127.0.0.1:8125`), prefix `openbank.lake` | `metrics.rs:39` |
| 4 | StatsD gauge: memory.bytes | StatsD gauge | `openbank.lake.memory.bytes` | `metrics.rs:55` |
| 5 | StatsD counter: message.relayed | StatsD count | `openbank.lake.message.relayed` | `metrics.rs:56-58` |
| 6 | systemd sd_notify | Unix datagram | `$NOTIFY_SOCKET` — sends `READY=1`, `STOPPING=1` | `program.rs:20,31` |
| 7 | Environment variable API | Environment | `LAKE_PORT_PULL`, `LAKE_PORT_PUB`, `LAKE_LOG_LEVEL`, `LAKE_STATSD_ENDPOINT` | `config.rs:19-22` |
| 8 | Config file | File | `/etc/lake/conf.d/init.conf` (EnvironmentFile for systemd) | `packaging/init.conf`, `packaging/debian/lake-relay.service:9` |
| 9 | Debian package | .deb | `lake_{version}_{arch}.deb`, installs to `/usr/bin/lake` | `packaging/debian/control`, `packaging/debian/lake.install.*` |
| 10 | Docker image | OCI | `openbank/lake:{arch}-{version}.{meta}` | `packaging/docker/*/Dockerfile` |
| 11 | CLI interface | Process | Binary `lake` — no CLI args, configured entirely via env vars, exits on SIGTERM | `main.rs` |

## Test architecture table (conversion)

| Aspect | Finding |
|--------|---------|
| Unit test framework | Rust built-in `cargo test` (no separate test files found in source — tests are implicit/minimal) |
| Blackbox test framework | Python 3 + `behave` (BDD), in `bbtest/` directory |
| Blackbox test runner | `bbtest/main.py` — runs behave, converts to cucumber JSON, then to JUnit XML |
| Blackbox test helpers | `helpers/unit.py` (systemd package install/configure/teardown), `helpers/zmq.py` (ZMQ PUSH/SUB test client), `helpers/eventually.py` (polling retry decorator), `helpers/logger.py` |
| Blackbox test steps | `steps/orchestration_steps.py` (install, systemctl, configure), `steps/zmq_steps.py` (send/receive messages), `steps/logs_steps.py` (journalctl assertions), `steps/metrics_steps.py` (StatsD mock assertions) |
| Blackbox test features | 6 features: install, configuration, management (start/stop/restart), metrics, relay (message ordering), uninstall |
| Blackbox test dependencies | `openbank_testkit` (Shell, Package, Platform, StatsdMock), `systemd.journal`, `zmq` (pyzmq), `behave`, `behave2cucumber` |
| Performance test framework | Python 3, in `perf/` directory |
| Performance test runner | `perf/main.py` — bootstraps appliance, pushes 1K→1B messages through ZMQ, collects metrics, generates graphs |
| Performance test publisher | `perf/messaging/publisher.py` — multiprocessing: one PUSH worker + one SUB worker |
| Performance benchmark | ~750,000 messages/sec throughput (amd64, 2GB RAM, 1 CPU) per README |
| Test execution environment | Docker container (`jancajthaml/bbtest:{arch}`) with systemd, docker socket mount, real package install |
| Test count | 0 Rust unit tests detected in source, 6 blackbox feature files with ~10 scenarios |

## Module dependency graph (conversion)

The crate is a single flat binary (no library crate, no sub-crates, no workspace). All modules are declared in `main.rs` via `mod` statements. The dependency graph is:

```
main.rs
├── config.rs       (no internal deps)
├── error.rs        (no internal deps — uses zmq_sys::errno)
├── logger.rs       (no internal deps — uses log, colored)
├── message.rs      (depends on: error [unused import?], zmq_sys)
├── metrics.rs      (depends on: config)
├── program.rs      (depends on: config, logger)
├── relay.rs        (depends on: config, error, message, metrics, socket)
└── socket.rs       (depends on: error, zmq_sys)
```

### Module roles and dependency counts

| Module | Dependants (used by) | Dependencies (uses) | Role | Leaf? |
|--------|---------------------|--------------------|----- |-------|
| `config` | `main`, `metrics`, `program`, `relay` | None (internal) | Configuration loading | Yes (leaf) |
| `error` | `relay`, `socket` | `zmq_sys` (external) | ZMQ error code mapping | Yes (leaf) |
| `logger` | `program` | `log`, `colored` (external) | Log implementation | Yes (leaf) |
| `message` | `relay` | `zmq_sys` (external) | ZMQ message wrapper | Yes (leaf) |
| `socket` | `relay` | `error`, `zmq_sys` (external) | ZMQ context/socket wrapper | Near-leaf (1 internal dep) |
| `metrics` | `main`, `relay` | `config`, `statsd` (external) | StatsD reporter | Near-leaf (1 internal dep) |
| `program` | `main` | `config`, `logger` | Daemon lifecycle | Near-leaf (2 internal deps) |
| `relay` | `main` | `config`, `error`, `message`, `metrics`, `socket` | Core relay loop | Core (5 internal deps) |
| `main` | None (entry point) | `config`, `metrics`, `program`, `relay` | Orchestration | Entry point |

No circular dependencies. `relay` is the core module with the most internal dependencies. `config`, `error`, `logger`, `message` are leaf modules.

## Raw notes

- The binary name in Cargo.toml is `main`, but the .deb packaging renames it to `lake` during `override_dh_installinit`.
- The relay loop is tight: `zmq_msg_recv` → `zmq_msg_send` with no intermediate processing, no filtering, no transformation — pure passthrough.
- ZMQ socket options are carefully tuned: `ZMQ_CONFLATE=0` (don't conflate messages), `ZMQ_IMMEDIATE=1` (only queue for completed connections), `ZMQ_LINGER=0` (don't block on close), `ZMQ_RCVHWM=0` / `ZMQ_SNDHWM=0` (unlimited high-water mark), `ZMQ_XPUB_NODROP=1` (error instead of dropping on slow subscribers).
- Shutdown uses an interesting pattern: the relay Drop impl sends a message to its own PULL socket via a temporary PUSH connection to unblock the blocking `zmq_msg_recv` call. Then the worker thread raises SIGTERM.
- The metrics thread also raises SIGTERM on exit — both worker threads use `libc::raise(libc::SIGTERM)` as a coordination mechanism to ensure the main thread's `sigwait` unblocks.
- Memory monitoring uses `/proc/self/stat` RSS via the `procfs` crate (Linux only). macOS implementation returns `0.0`.
- The logger implements manual UTC date formatting from UNIX epoch — does not use the `chrono` crate at runtime (chrono is a transitive dependency of `procfs` but not used directly).
- There is a typo in `program.rs:55`: `fn notify(msg: &msg)` should be `fn notify(msg: &str)` — dead code on non-unix, never compiled.
- The `zmq-sys` crate links against system `libzmq` via `pkg-config` (the `metadeps` crate). This is a dynamic link — the binary requires `libzmq5` at runtime.
- Performance is the primary design concern: the README advertises 750K msg/sec on modest hardware.
- The project belongs to the OpenBank platform — the StatsD prefix is `openbank.lake`, suggesting this is one component among several (`vault`, etc.) that communicate via this relay.
