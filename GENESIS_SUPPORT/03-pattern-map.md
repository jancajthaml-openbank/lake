# Pattern Map — Lake

Rust (edition 2021, with unsafe FFI to C via `zmq-sys`/`libc` crates) → AArch64 (arm64) assembly (GNU `as` syntax, targeting Linux, linking dynamically against `libzmq.so` and `libc.so`). This translation is non-trivial because Rust provides compile-time memory safety guarantees (ownership, borrowing, lifetimes), a type system with algebraic data types (enums, Result, Option), automatic resource cleanup (Drop trait), and a standard library with managed strings, threading, and atomic types — none of which exist in raw assembly. However, the codebase already operates almost entirely through `unsafe` FFI blocks calling C library functions, which significantly reduces the semantic gap: the Rust code is already closer to C-with-safety-wrappers than to idiomatic safe Rust.

## Pattern inventory

| # | Source pattern | Idiom / example | Frequency | Structural | Risk |
|---|---------------|----------------|-----------|-----------|------|
| 1 | Unsafe FFI calls to C library (libzmq) | `zmq_sys::zmq_msg_recv(ptr, puller.sock, 0_i32)` inside `unsafe {}` | 31 instances / 6 files | Yes | Low |
| 2 | Raw C pointer management (`*mut c_void`) | `pub underlying: *mut c_void` as ZMQ handle | 2 structs, pervasive | Yes | Low |
| 3 | Manual `Drop` implementations (RAII cleanup) | `impl Drop for Socket { fn drop(&mut self) { zmq_close(self.sock) } }` | 5 implementations | Yes | Medium |
| 4 | `unsafe impl Send + Sync` for thread-safety marker | `unsafe impl Send for Context {}` | 1 instance | Yes | Low |
| 5 | `MaybeUninit` + `assume_init` for deferred init | `let mut sigset_memspace = MaybeUninit::uninit(); ... sigset_memspace.assume_init()` | 2 instances | Yes | Low |
| 6 | `extern "C" fn` callback for signal handler | `extern "C" fn handler(_: libc::c_int) {}` | 1 instance | Yes | Low |
| 7 | POSIX signal API via libc crate | `libc::signal(SIGTERM, blocking)`, `libc::sigwait(...)` | 4 calls | Yes | Low |
| 8 | `libc::raise(SIGTERM)` for thread-to-main communication | Child thread raises SIGTERM after loop exit to unblock main | 2 instances | Yes | Low |
| 9 | Conditional compilation `#[cfg(target_os)]` | `#[cfg(target_os = "linux")] fn mem_bytes()` and `#[cfg(not(target_family = "unix"))]` | 4 cfg blocks | Medium | Low |
| 10 | Custom logger with manual date/time calculation | `let dayclock = ts % 86400;` — avoids chrono at runtime, computes Y/M/D from epoch | 1 impl (~50 lines) | Medium | Medium |
| 11 | `Arc<AtomicBool>` cross-thread cancellation flag | `prog.running.clone().store(false, Ordering::Relaxed)` checked in relay and metrics loops | Pervasive (4 files) | Yes | Medium |
| 12 | `Arc<AtomicUsize>` lock-free accumulator | `self.messages.fetch_add(1, Ordering::Relaxed)` with periodic `swap(0, ...)` drain | 1 instance | Medium | Low |
| 13 | `std::thread::spawn` for OS-level concurrency | `thread::spawn(move \|\| { ... })` for relay loop and metrics reporter | 2 threads | Yes | High |
| 14 | `ffi::CString` for C string interop | `CString::new(endpoint.as_bytes()).unwrap()` before passing to `zmq_bind` | 2 instances | Medium | Low |
| 15 | `mem::transmute` for lifetime extension | `let v: &'static [u8] = mem::transmute(ffi::CStr::from_ptr(s).to_bytes())` | 1 instance | Low | Low |
| 16 | Rust `Result<T, E>` for error propagation | `fn bind(&self, endpoint: &str) -> Result<(), error::Error>` | ~15 return sites | Yes | Medium |
| 17 | Rust `match` exhaustive pattern matching | `match self { Error::EACCES => ..., Error::EADDRINUSE => ..., ... }` covering 28 errno variants | ~6 match blocks | Yes | Medium |
| 18 | Rust `String` / `format!` macro for dynamic strings | `format!("tcp://0.0.0.0:{}", port)` for endpoint construction | ~8 instances | Medium | Medium |
| 19 | `statsd::Client` UDP pipeline pattern | `let mut pipe = statsd_client.pipeline(); pipe.gauge(...); pipe.count(...); pipe.send(...)` | 1 usage | Medium | High |
| 20 | `procfs::process::Process::myself().stat().rss` | Linux-specific RSS memory via procfs crate abstractions | 1 usage | Low | Medium |
| 21 | Rust `println!` / `log::info!` macro expansion | Compile-time string formatting + runtime `Log::log()` dispatch | All files | Medium | Medium |
| 22 | Closure capture in `thread::spawn(move \|\| {...})` | Moves `ctx`, `pull_port`, `pub_port`, `prog_running`, `metrics` into thread | 2 instances | Yes | High |

## Pattern mapping

| # | Source pattern | Target equivalent | Semantic gap | Mitigation | ADR |
|---|---------------|------------------|-------------|------------|-----|
| 1 | Unsafe FFI calls to libzmq | `BL zmq_msg_recv` / `BL zmq_msg_send` — direct C ABI function calls via PLT | None — Rust unsafe FFI already generates identical call sequences | — | TBD |
| 2 | Raw C pointer management (`*mut c_void`) | Register-based pointer management (e.g. `X19` holds zmq context pointer) | None — assembly is already raw pointer manipulation | — | TBD |
| 3 | Manual `Drop` (RAII cleanup) | Explicit cleanup code blocks at every exit path; structured as labeled branch targets (e.g. `.Lcleanup_socket:`) | Compiler no longer ensures cleanup runs on all paths — a missed branch skips cleanup and leaks resources | Every function with resources must use a single-exit cleanup pattern; review all branch paths manually | TBD |
| 4 | `unsafe impl Send + Sync` | Not applicable — no type system in assembly; thread safety is entirely the programmer's responsibility | The compile-time guarantee that data is safe to share across threads disappears | Document which registers/memory hold shared state; use appropriate memory barriers | TBD |
| 5 | `MaybeUninit` + `assume_init` | Stack-allocated uninitialized memory (`SUB SP, SP, #size`) used directly | None — assembly stack allocation is inherently uninitialized | — | TBD |
| 6 | `extern "C" fn` signal handler | Label in `.text` section conforming to C calling convention (`X0` = signum) | None — direct equivalent | — | TBD |
| 7 | POSIX signal API via libc | `BL sigaction` / `SVC #0` with `rt_sigaction` syscall number | None substantive — same libc functions or direct syscalls | Choose between libc calls (simpler) or raw syscalls (no libc dependency for signal path) | TBD |
| 8 | `libc::raise(SIGTERM)` | `BL raise` or `SVC #0` with `kill` syscall (`getpid` + `kill(pid, SIGTERM)`) | None — direct equivalent | — | TBD |
| 9 | Conditional compilation `#[cfg]` | Not applicable for single-target arm64 build — linux/arm64 is the only target; dead code eliminated at source level | Source must be written for linux/arm64 only; no macOS `mem_bytes` stub needed | Remove macOS/non-unix code paths entirely; single-platform assembly | TBD |
| 10 | Custom logger (manual date calc) | Assembly subroutine: `clock_gettime(CLOCK_REALTIME)` → manual epoch-to-YMD conversion → `write(1, buf, len)` | Semantic parity achievable but labour-intensive; ~100 instructions for date formatting | Reimplement the same arithmetic (ts%86400 for time-of-day, year/month loop) in assembly | TBD |
| 11 | `Arc<AtomicBool>` cross-thread flag | Shared memory location (global `.bss` or heap) accessed with `LDAR`/`STLR` (acquire/release) instructions | Rust `Arc` provides reference-counted shared ownership; assembly uses a global with known lifetime (program duration) — simpler since the flag outlives all threads | Use `.bss` global; no reference counting needed because lifetime is static | TBD |
| 12 | `Arc<AtomicUsize>` lock-free counter | Shared `.bss` global accessed with `LDXR`/`STXR` (exclusive) for `fetch_add`, `LDAR`/`STLR` for `swap` | Same as #11 — global lifetime simplifies away `Arc` | Use `LDAXR`/`STLXR` loop for atomic increment; `SWP` or `LDXR`/`STXR` for atomic swap-to-zero | TBD |
| 13 | `std::thread::spawn` | `BL pthread_create` with function pointer + argument pointer, or `clone` syscall with new stack | Rust handles stack allocation, panic unwinding, and thread-local storage automatically; assembly must manually allocate stack (`mmap`), set up function pointer, and clean up on exit | Allocate thread stacks via `mmap(PROT_READ\|PROT_WRITE, MAP_PRIVATE\|MAP_ANONYMOUS\|MAP_STACK)`; use `pthread_create` for simplicity over raw `clone` | TBD |
| 14 | `ffi::CString` (null-terminated C string) | Null-terminated string literals in `.rodata` section, or stack buffer with null terminator for dynamic strings | None for static strings; for dynamic strings (format with port number), must manually write digits + null terminator into stack buffer | Integer-to-ASCII subroutine for port number formatting | TBD |
| 15 | `mem::transmute` lifetime extension | Direct pointer use — no lifetime concept in assembly | None — assembly pointers have no lifetime annotations | — | TBD |
| 16 | `Result<T, E>` error propagation | Register-based convention: `X0` = return value (0 = success, or result), `W0` = -1 on error; check with `CBZ`/`CBNZ`/`CMP` | No compiler enforcement of error checking — a missed check silently ignores errors | Establish and document a calling convention; check return values at every call site | TBD |
| 17 | `match` exhaustive pattern matching | `CMP`/`B.EQ` chains or jump table (`ADR` + `BR`) for errno dispatch | Compiler no longer verifies exhaustiveness — a missing case falls through silently | Use jump table with default branch for unknown codes; test all errno paths | TBD |
| 18 | `String` / `format!` dynamic strings | Fixed-size stack buffers + manual `snprintf`-style assembly (digit conversion, `STP`/`STR` to buffer) | Rust `String` grows dynamically; assembly uses fixed buffers that could overflow | Size buffers conservatively (e.g. 128 bytes for `tcp://0.0.0.0:65535\0` is 24 bytes max); assert no overflow | TBD |

## Unmappable patterns

| # | Source pattern | Frequency | Why unmappable | Proposed workaround | Impact |
|---|---------------|-----------|---------------|--------------------|---------| 
| U1 | Rust ownership/borrowing/lifetime system | Pervasive (every variable binding) | No assembly equivalent — this is a compile-time analysis with no runtime representation. There is no instruction or mechanism that enforces ownership rules. | Programmer discipline: document ownership of every pointer/register; use single-exit cleanup patterns; never alias mutable state across threads without atomics. No automated verification possible. | **High** — All memory safety guarantees are lost. Use-after-free, double-free, and data races become possible. Mitigated by the fact that the program is small (~500 lines) and the ownership graph is simple (no complex lifetimes). |
| U2 | Rust type system (enums with data, generics, traits) | ~10 type definitions, `Error` enum (28 variants), `Log` trait impl, `Result<T,E>`, `Option<T>` | No assembly equivalent — types exist only at compile time. Assembly operates on untyped registers and memory. | Establish register/memory layout conventions and document them. Use integer codes for error variants (matching `zmq_errno()` values directly, eliminating the enum translation layer). | **Medium** — Loss of type safety means type confusion bugs are possible. Mitigated by the program's simplicity: only a few "types" exist (ZMQ context pointer, socket pointer, message struct, integer config values). |
| U3 | Cargo build system + crate ecosystem | 1 Cargo.toml, 6 crate dependencies | No assembly equivalent. Cargo resolves dependencies, manages compilation units, runs tests. Assembly has no package manager or dependency resolution. | Replace with Makefile + GNU `as` + `ld`. All crate functionality (statsd, colored, procfs, log) must be reimplemented in assembly. The only external dependency retained is `libzmq.so` (linked dynamically). | **High** — All Rust library code must be hand-written in assembly. The `statsd` crate (~UDP socket + protocol formatting), `colored` (~ANSI escapes), `procfs` (~/proc parsing), and `log` (~function dispatch) must all be reimplemented. Estimated ~400-600 instructions of new code. |
| U4 | Rust standard library (`std::thread`, `std::sync::Arc`, `std::env`, `std::io`, `std::ffi`, `std::time`) | Used in every module | No assembly equivalent. The Rust standard library provides platform-abstracted threading, atomics, I/O, environment access, and string handling. | Replace each `std` usage with direct Linux arm64 syscalls or libc C ABI calls: `pthread_create` for threads, `getenv` for env vars, `socket`/`sendto` for UDP, `write` for stdout, `clock_gettime` for time, `open`/`read`/`close` for /proc. | **High** — Every standard library call must be replaced. However, the codebase already uses `libc`/`zmq-sys` for most operations, so the actual unique `std` functionality to reimplement is limited: `thread::spawn` (→ `pthread_create`), `env::var_os` (→ `getenv`), `UnixDatagram` (→ `socket`/`sendto`), `SystemTime` (→ `clock_gettime`). |
| U5 | Closure capture with `move` semantics in `thread::spawn` | 2 instances (relay thread, metrics thread) | Closures are a compiler-generated anonymous struct containing captured variables, with a generated `call` method. No assembly equivalent. | Pass thread arguments via a struct pointer allocated on the heap (or in `.bss` for program-lifetime data). `pthread_create` accepts a `void* arg` parameter — pack captured values into a struct and pass its address. | **Medium** — Requires manual struct layout for thread arguments. The relay thread needs: `ctx`, `pull_port`, `pub_port`, `prog_running`, `metrics_ptr`. The metrics thread needs: `endpoint_ptr`, `messages_ptr`, `prog_running`. Both are small, fixed sets. |
| U6 | `statsd::Client` crate (UDP StatsD protocol) | 1 crate, used in metrics thread | No assembly equivalent crate. The `statsd` crate creates a UDP socket, formats StatsD protocol lines (`metric.name:value|type`), batches them in a pipeline, and sends. | Reimplement in assembly: `socket(AF_INET, SOCK_DGRAM, 0)` → `connect` to endpoint → format metric strings into buffer → `sendto`. The StatsD text protocol is trivial: `openbank.lake.message.relayed:N|c\nopenbank.lake.memory.bytes:N|g\n`. Estimated ~80-120 instructions. | **Medium** — The StatsD protocol is simple enough to reimplement. Risk is in endpoint string parsing (`host:port` → `sockaddr_in`). |
| U7 | `procfs` crate (Linux /proc/self/stat parsing) | 1 crate, 1 call site | No assembly equivalent. The crate reads `/proc/self/stat`, parses the whitespace-delimited fields, and returns structured data. | Reimplement in assembly: `open("/proc/self/stat", O_RDONLY)` → `read` into buffer → scan for 24th space-delimited field (RSS) → multiply by page size (`getauxval(AT_PAGESZ)` or hardcode 4096). Estimated ~60-80 instructions. | **Low** — The parsing is mechanical: skip N fields, read one integer. Single call site. |

## External contract inventory

| # | Contract | Type | Endpoint / schema | Must preserve |
|---|----------|------|------------------|--------------|
| 1 | ZMQ PULL socket (message ingress) | ZMTP 3.0 / TCP | `tcp://0.0.0.0:{LAKE_PORT_PULL}` (default 5562) | Yes |
| 2 | ZMQ PUB socket (message egress) | ZMTP 3.0 / TCP | `tcp://0.0.0.0:{LAKE_PORT_PUB}` (default 5561) | Yes |
| 3 | ZMQ PUSH poison-pill (shutdown) | ZMTP 3.0 / TCP | `tcp://127.0.0.1:{LAKE_PORT_PULL}` (loopback) | Yes |
| 4 | StatsD metric: `openbank.lake.message.relayed` | UDP / StatsD count | `{LAKE_STATSD_ENDPOINT}` (default 127.0.0.1:8125) | Yes |
| 5 | StatsD metric: `openbank.lake.memory.bytes` | UDP / StatsD gauge | `{LAKE_STATSD_ENDPOINT}` (default 127.0.0.1:8125) | Yes |
| 6 | Systemd notification: `READY=1` | Unix datagram | `$NOTIFY_SOCKET` | Yes |
| 7 | Systemd notification: `STOPPING=1` | Unix datagram | `$NOTIFY_SOCKET` | Yes |
| 8 | Environment variables: `LAKE_PORT_PULL`, `LAKE_PORT_PUB`, `LAKE_LOG_LEVEL`, `LAKE_STATSD_ENDPOINT` | Process environment | Read at startup via `getenv` | Yes |
| 9 | Binary interface: `/usr/bin/lake` (no arguments) | ELF executable | Installed via .deb, started by systemd | Yes |
| 10 | Exit behaviour: clean shutdown on SIGTERM | POSIX signal | `SIGTERM` → stop relay → close sockets → exit 0 | Yes |
| 11 | Log output format: `YYYY-MM-DDThh:mm:ssZ LVL [target] message` | stdout | Written to fd 1 | Yes |
