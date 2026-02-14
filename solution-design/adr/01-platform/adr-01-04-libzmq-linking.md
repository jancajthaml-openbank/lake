# ADR-01-04: libzmq linking strategy — dynamic linking

## Context

The relay depends on the libzmq C library. The current Rust build dynamically links against `libzmq5 >= 4.3, < 4.4` declared in the Debian control file. Zig's build system supports both dynamic and static linking. The choice affects the `.deb` package dependencies, the Docker image size, and failure scenario 5 (libzmq5 not installed).

See also: [PRD-004](../../prd/prd-004-deployment-and-distribution.md), [PRD-005](../../prd/prd-005-conversion-contract-preservation.md)

## Decision

We choose **dynamic linking** against system `libzmq5`, matching the current behaviour.

Rationale:
- Preserves the existing `.deb` dependency declaration (`Depends: libzmq5 (>= 4.3~), libzmq5 (<< 4.4~)`) — no packaging changes needed (invariant 5).
- Existing blackbox tests install the `.deb` which pulls in libzmq5 as a dependency — changing to static linking would alter this flow and could break tests (invariant 6).
- The operator already has libzmq5 installed (other OpenBank services likely depend on it too).
- Dynamic linking allows the operator to patch libzmq5 independently of the lake binary (security updates).

The `build.zig` will use `exe.linkSystemLibrary("zmq")` which instructs the linker to link against `-lzmq` at build time and `libzmq5.so` at runtime.

### Alternatives considered

| Alternative | Why rejected |
|------------|-------------|
| Static linking (bundle libzmq into binary) | Would eliminate failure scenario 5 but changes the `.deb` dependency declaration, increases binary size (~2-3 MB), and prevents independent libzmq patching. Could be offered as a build flag for container deployments in the future but not as the default. |
| Vendored source build of libzmq | Massive build complexity (libzmq is a large C++ project). Zig can compile C but libzmq has C++ components. Not worth the effort for a well-packaged system library. |

## Failure modes and mitigations

| Failure mode | Mitigation |
|-------------|-----------|
| `libzmq5.so` not installed at runtime | The `.deb` declares the dependency — `apt` will install it. If manually installing the binary without the `.deb`, the operator must install `libzmq5` separately. The binary will fail immediately with a clear dynamic linker error. |
| libzmq5 version mismatch (< 4.3 or ≥ 4.4) | The `.deb` dependency pins `>= 4.3~, << 4.4~`. The `@cImport` compiles against headers from the build environment — ensure CI uses the same libzmq-dev version range. |
| libzmq5 ABI break within the 4.3.x series | Unlikely — libzmq follows semver for the C ABI. Pin the `.deb` dependency range. If a break occurs, the blackbox tests will catch it. |
| Cross-compilation cannot find target-arch libzmq5 | CI must provide cross-architecture library paths. Zig's `addLibraryPath` can point to the target-arch sysroot. |

## Consequences

- The `.deb` control file retains `Depends: libzmq5 (>= 4.3~), libzmq5 (<< 4.4~)`.
- Runtime component count remains 2 (binary + libzmq5.so), within the component budget.
- The Docker image continues to install libzmq5 via the `.deb` dependency.
- Static linking remains available as a future option via a `build.zig` flag, but is not the default.
