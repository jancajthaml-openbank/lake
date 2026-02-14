# ADR-01-01: Language — Zig

## Context

The lake relay is currently implemented in Rust 2021. The user's architectural intent is to convert to Zig for improved reliability (eliminating unsafe FFI indirection) and performance (reducing allocation overhead, leveraging direct C interop). The relay is ~400 lines across 9 modules with 31 `unsafe` blocks, 6 `Drop` impls, and pervasive `Arc`-based shared state — all patterns that translate differently in Zig.

See also: [PRD-001](../../prd/prd-001-high-throughput-message-relay.md), [PRD-005](../../prd/prd-005-conversion-contract-preservation.md), [PRD-006](../../prd/prd-006-build-and-compilation.md)

## Decision

We choose **Zig** (pinned to a specific stable release, e.g., 0.13.x) as the implementation language.

Rationale:
- Zig's `@cImport` provides zero-cost C interop — ZMQ C calls are made directly without an FFI binding layer, eliminating the `zmq-sys` crate and its `unsafe` wrapper overhead.
- Zig has no hidden allocations, no runtime, and no garbage collector — the hot relay loop (recv → send) can be written with zero per-message heap allocations.
- Zig's built-in cross-compilation (`-target x86_64-linux`, `-target aarch64-linux`) is simpler than the current Cargo + clang-13 LTO setup.
- Zig's `defer` replaces Rust's `Drop` for resource cleanup, making cleanup order explicit and auditable.

The Zig version must be pinned in `build.zig.zon` (or equivalent) and in CI configuration to avoid breakage from pre-1.0 language changes.

### Alternatives considered

| Alternative | Why rejected |
|------------|-------------|
| Keep Rust | Does not satisfy user's stated conversion intent |
| C | No safety improvements; Zig provides comparable performance with better error handling and build system |
| C++ | Higher complexity; no `@cImport` equivalent; complex build tooling |
| Go | Runtime and GC would introduce latency jitter in the hot relay loop; violates zero-allocation requirement |

## Failure modes and mitigations

| Failure mode | Mitigation |
|-------------|-----------|
| Zig compiler breaking changes between versions (pre-1.0 instability) | Pin exact Zig version in build definition and CI. Test build on version upgrade before adopting. |
| Loss of Rust's compile-time safety (borrow checker, Send/Sync) | Codebase is ~400 lines with a simple linear lifecycle. Use `defer` for all cleanup, atomics for all shared state. Blackbox tests provide behavioural safety net. Use Zig's `GeneralPurposeAllocator` with leak detection in test builds. |
| Zig ecosystem immaturity (fewer libraries, less tooling) | The binary has exactly one external C dependency (libzmq). No Zig-native libraries needed. StatsD client and `/proc` reader are hand-rolled (~40 lines combined). |
| Developer unfamiliarity with Zig | Codebase is small. Zig's C interop means the ZMQ layer reads like C code. Pattern map documents every Rust→Zig translation. |

## Consequences

- All Rust source (`services/lake/src/`) is replaced with Zig source.
- `Cargo.toml`, `Cargo.lock`, `.rustfmt.toml` are removed.
- All Rust dependencies (`zmq-sys`, `statsd`, `log`, `libc`, `colored`, `procfs`) are eliminated. The only external dependency is the libzmq C library.
- CI images change from `jancajthaml/rust:{arch}` to a Zig-based image.
- Developers must have the Zig compiler installed for local builds.
- See [ADR-01-02](adr-01-02-build-system.md) for build system consequences.
