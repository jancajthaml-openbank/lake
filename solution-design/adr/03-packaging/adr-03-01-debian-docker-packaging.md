# ADR-03-01: Debian and Docker packaging — preserve existing structure

## Context

The relay is deployed as a `.deb` package (installs binary, systemd units, default config) and as Docker images (Debian sid-slim installing the `.deb`). Existing automation depends on the package name (`lake`), binary path (`/usr/bin/lake`), systemd unit names, and Docker image tags. The conversion must produce identical packaging artifacts from the Zig build instead of the Rust build.

See also: [PRD-004](../../prd/prd-004-deployment-and-distribution.md), [PRD-005](../../prd/prd-005-conversion-contract-preservation.md)

## Decision

The existing `packaging/` directory structure, Debian control files, systemd unit files, Dockerfiles, and `init.conf` are **preserved unchanged** (or with minimal changes to binary source paths).

Changes required:
1. `packaging/debian/control` — `Build-Depends` line updated (remove Rust toolchain, note Zig if needed). `Depends` line unchanged: `libzmq5 (>= 4.3~), libzmq5 (<< 4.4~)`.
2. `packaging/debian/lake.install.amd64` — update binary source path from `bin/lake-linux-amd64` to the Zig build output path (e.g., `bin/lake-linux-amd64`). If the `dev/lifecycle/package` script copies the Zig output to the same path, no change needed.
3. `packaging/debian/lake.install.arm64` — same as above for arm64.
4. `packaging/docker/amd64/Dockerfile` and `packaging/docker/arm64/Dockerfile` — unchanged (they install the `.deb`, which installs the binary).
5. `packaging/debian/rules` — the `override_dh_installinit` step that renames the binary to `lake` must work with the Zig output. The Zig build should produce a binary named `lake-linux-{arch}` to match the expected pattern, or the rules file should be updated to match the Zig output name.
6. All systemd unit files (`lake.service`, `lake-relay.service`, `lake-watcher.path`, `lake-watcher.service`) — **unchanged**.
7. `packaging/init.conf` — **unchanged**.

### Alternatives considered

| Alternative | Why rejected |
|------------|-------------|
| Replace `.deb` packaging with raw binary distribution | Breaks the operator workflow, loses systemd unit installation, loses dependency management. Violates PRD-004 and invariant 5. |
| Replace Docker images with a scratch-based image (no OS) | Would require static linking (see [ADR-01-04](../01-platform/adr-01-04-libzmq-linking.md)). The current approach installs via `.deb` inside Debian — changing this breaks the packaging contract. |
| Create new packaging format (Snap, Flatpak, AppImage) | No demand, increases operator burden, violates the minimal operator profile constraint. |

## Failure modes and mitigations

| Failure mode | Mitigation |
|-------------|-----------|
| Zig binary name does not match the pattern expected by `debian/rules` | The `dev/lifecycle/package` script must copy the Zig build output to `packaging/bin/lake-linux-{arch}` — the same location the Rust build used. The `rules` file's `override_dh_installinit` renames `lake-*` to `lake`. |
| `.deb` package installs but binary fails to start (missing dynamic library, wrong arch) | The blackbox `install.feature` test installs the package and verifies all systemd units are active. This catches any packaging error (invariant 6). |
| Docker image build fails (`.deb` not found in expected path) | The CI pipeline produces the `.deb` before building the Docker image. The path `packaging/bin/` is the staging area for both Rust and Zig builds. |
| systemd unit incompatibility with Zig binary (e.g., sd_notify timing) | The systemd units are unchanged. The Zig binary must satisfy the same contracts: `Type=notify`, exit on SIGTERM, etc. Verified by `management.feature` and `install.feature` tests. |

## Consequences

- The `packaging/` directory is largely unchanged — the conversion is invisible to the packaging layer.
- The `dev/lifecycle/package` script is the integration point: it changes from calling `cargo build` to calling `zig build`, but produces output at the same path.
- The Docker images, `.deb` structure, systemd units, and `init.conf` are identical before and after conversion.
- Interacts with [ADR-01-02](../01-platform/adr-01-02-build-system.md) (build produces binary at expected path), [ADR-01-04](../01-platform/adr-01-04-libzmq-linking.md) (dynamic linking preserves `.deb` dependency).
