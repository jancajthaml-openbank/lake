# ADR-03-02: CI/CD pipeline — adapt existing pipelines for Zig

## Context

The project uses three CI systems: CircleCI (primary, with full pipeline including blackbox tests), GitHub Actions (health check build, DevSkim scanner), and Jenkins (alternative pipeline). All use Docker images with the Rust toolchain (`jancajthaml/rust:{arch}`). The conversion requires replacing the Rust build environment with a Zig build environment while preserving the pipeline structure: build → test → package-debian → package-docker → blackbox-test → publish → release.

See also: [PRD-006](../../prd/prd-006-build-and-compilation.md), [PRD-005](../../prd/prd-005-conversion-contract-preservation.md)

## Decision

The existing CI pipeline structure is preserved. Only the build environment and build commands change:

1. **CI Docker image**: Replace `jancajthaml/rust:{arch}` with a Zig-based image (e.g., `jancajthaml/zig:{arch}` or a public Zig image). The image must contain: Zig compiler (pinned version), libzmq-dev headers, and standard build tools. For cross-compilation, the image must contain target-architecture libzmq libraries.

2. **Build commands**: Replace `cargo build --release --target=...` with `zig build -Doptimize=ReleaseFast -Dtarget=...`. Replace `cargo test` with `zig build test`. Replace `cargo fmt && cargo clippy` with `zig fmt`.

3. **Lifecycle scripts**: Update `dev/lifecycle/package`, `dev/lifecycle/test`, `dev/lifecycle/lint`, `dev/lifecycle/sync` to call Zig equivalents. The script interface (`--source`, `--output`, `--arch`) remains unchanged so the CI config files need minimal changes.

4. **CircleCI config**: Update the `rust` executor to a Zig-based executor. Update the `unit-test`, `compile`, and `lint`/`sec` jobs. The `package-debian`, `package-docker`, `blackbox-test`, `publish`, and `release` jobs remain unchanged (they operate on build artifacts, not source code).

5. **GitHub Actions health check**: Update to install Zig instead of Rust, install libzmq-dev, run `zig build` and `zig build test`.

6. **Dependabot**: Remove `.github/dependabot.yml` (Cargo ecosystem). No Zig equivalent needed — the binary has no Zig package registry dependencies.

7. **Zig version pinning**: The CI image Dockerfile must pin the exact Zig version. Do not use `latest`. Zig version upgrades are explicit, deliberate, and tested.

### Alternatives considered

| Alternative | Why rejected |
|------------|-------------|
| Rewrite CI pipelines from scratch | Unnecessary — the pipeline structure (build → test → package → test → publish) is sound. Only the build tool changes. |
| Use GitHub Actions as the primary CI | The project's primary CI is CircleCI with ARM machine executors for arm64 builds and blackbox testing. GitHub Actions lacks native arm64 runners. Keep CircleCI as primary. |
| Remove Jenkins pipeline | Out of scope for this conversion. The Jenkins pipeline may still be used in some environments. Update it alongside CircleCI. |

## Failure modes and mitigations

| Failure mode | Mitigation |
|-------------|-----------|
| Zig CI image not available for arm64 | Build a custom Docker image with Zig for arm64. Zig provides official tarballs for aarch64-linux — the Dockerfile downloads and installs the pinned version. |
| CI cache invalidation after Zig migration | Zig uses a different cache structure than Cargo. Update CircleCI cache keys from `code-*` to include a Zig-specific prefix. The global Zig cache directory (`~/.cache/zig`) can be cached. |
| `dev/lifecycle/` script changes break CI | The script interface (`--source`, `--output`, `--arch` flags) is preserved. Only the internal commands change. Test the scripts locally before merging CI changes. |
| Zig version drift between developer machines and CI | Pin the version in `build.zig.zon` and in the CI Dockerfile. Zig's `build.zig.zon` can specify a minimum version. CI fails fast if the wrong Zig version is used. |
| Blackbox tests fail after CI migration | The blackbox tests are language-agnostic — they test the `.deb` package and the binary's external behaviour. If they fail, the issue is in the build output (wrong binary name, missing library), not in the tests. CI must run blackbox tests as a gate before publish. |

## Consequences

- CI pipeline structure unchanged — same stages, same artifact flow.
- Build environment changes: Rust image → Zig image. The image must be created and maintained.
- `dev/lifecycle/` scripts updated with Zig commands but same external interface.
- `.github/dependabot.yml` removed (Cargo-specific).
- DevSkim scanner (GitHub Actions) continues to work — it scans source files regardless of language.
- The rolling hourly blackbox contract test on CircleCI continues unchanged — it installs the `.deb` and runs tests.
- Interacts with [ADR-01-01](../01-platform/adr-01-01-language-zig.md) (language choice), [ADR-01-02](../01-platform/adr-01-02-build-system.md) (build commands), [ADR-03-01](adr-03-01-debian-docker-packaging.md) (package artifacts).
