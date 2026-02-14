# ADR-01-02: Build system — Zig build

## Context

The current Rust build uses Cargo with a single crate, cross-compiled via `cargo build --target` with clang-13 as the linker and LTO enabled. The Zig conversion requires a new build system. The build must support dual-arch cross-compilation (amd64, arm64), release optimization, binary stripping, and integration with the existing `dev/lifecycle/` shell scripts.

See also: [PRD-006](../../prd/prd-006-build-and-compilation.md), [ADR-01-01](adr-01-01-language-zig.md)

## Decision

We use Zig's native build system via a `build.zig` file placed at `services/lake/build.zig`.

The build definition will:
- Compile the lake binary with `ReleaseFast` optimization (equivalent to Rust's `opt-level = 3`).
- Link against libzmq via `exe.linkSystemLibrary("zmq")` (dynamic) or `exe.addObjectFile(...)` (static) — see [ADR-01-03](adr-01-03-zmq-binding.md).
- Support cross-compilation via `zig build -Dtarget=aarch64-linux` — no external linker required.
- Strip the binary in release mode (`exe.root_module.strip = true`).
- Pin the minimum Zig version in `build.zig.zon`.

Developer workflow: `zig build` for debug, `zig build -Doptimize=ReleaseFast` for release, `zig build test` for unit tests.

### Alternatives considered

| Alternative | Why rejected |
|------------|-------------|
| CMake wrapping Zig sources | Unnecessary complexity; Zig's build system is native and sufficient |
| Make calling `zig build-exe` directly | Loses Zig build system features (dependency tracking, caching, cross-compilation flags) |
| Keep Cargo with Zig as a C dependency | Inverts the conversion — Zig should be the primary language, not a C library consumed by Rust |

## Failure modes and mitigations

| Failure mode | Mitigation |
|-------------|-----------|
| `build.zig` API changes in future Zig versions | Pin Zig version. The build file is ~30 lines; updating it for a new Zig release is a small effort. |
| Cross-compilation fails to find libzmq headers/library for target arch | Use Zig's `addSystemIncludePath` and `addLibraryPath` with explicit paths per target. CI images must have both amd64 and arm64 libzmq-dev installed or cross-libs available. |
| `dev/lifecycle/` scripts assume Cargo commands | Update `dev/lifecycle/package`, `dev/lifecycle/test`, `dev/lifecycle/lint` to call `zig build` instead of `cargo build`. The script interface (flags `--source`, `--output`, `--arch`) remains the same. |
| Build cache invalidation issues | Zig's incremental compilation cache is deterministic. `zig build --release=fast` always produces the same output for the same input. CI can cache the Zig global cache directory. |

## Consequences

- `Cargo.toml`, `Cargo.lock` are deleted.
- `build.zig` and `build.zig.zon` are created at `services/lake/`.
- `dev/lifecycle/package` script is updated: replaces `cargo build --release --target=...` with `zig build -Doptimize=ReleaseFast -Dtarget=...`.
- `dev/lifecycle/test` script is updated: replaces `cargo test` with `zig build test`.
- `dev/lifecycle/lint` script is updated: replaces `cargo fmt && cargo clippy` with `zig fmt`.
- No external build dependencies beyond the Zig compiler and libzmq-dev headers.
- Cross-compilation no longer requires clang-13, LLD, or separate toolchain installation.
