# PRD-004: Deployment and distribution

## Problem / requirement

The relay must be installable on Debian-family Linux hosts via a `.deb` package and runnable as a Docker container. The operator's deployment workflow is: install the package (which installs the binary, the systemd units, and the default configuration), then start the service. Existing automation, configuration management scripts, and container orchestration depend on the current package name, binary path, systemd unit names, and Docker image tags.

The conversion must not change the operator-facing deployment interface. The package must be the same name, the binary must install to the same path, the systemd units must have the same names and the same behaviour. The Docker image must be tagged with the same naming convention. The operator who upgrades from the Rust version to the Zig version must not need to change any deployment automation.

The system must be buildable for both amd64 and arm64 Linux. The current build uses separate CI matrix jobs for each architecture. The Zig build must support both targets.

The system must not introduce additional runtime components. The operator manages one binary and one shared library (libzmq5). The conversion must not add daemons, sidecars, agents, or runtime dependencies beyond what exists today.

## Success criteria

- The build must produce a `.deb` package named `lake_{version}_{arch}.deb` that installs the binary to `/usr/bin/lake`.
- The `.deb` package must include the systemd units `lake.service`, `lake-relay.service`, `lake-watcher.path`, and `lake-watcher.service` with identical behaviour to the current units.
- The `.deb` package must include the default configuration file at `/etc/lake/conf.d/init.conf`.
- The `.deb` package must declare a dependency on `libzmq5` (or, if static linking is chosen, must not require libzmq5 at runtime and the dependency must be removed from the control file).
- The build must produce Docker images tagged `openbank/lake:{arch}-{version}.{meta}` for both amd64 and arm64.
- The build must produce binaries for both amd64 and arm64 Linux from a single build definition.
- The total count of runtime components must not exceed 2 (the binary plus optionally the libzmq5 shared library).

## Out of scope

- RPM packaging or non-Debian distributions.
- Kubernetes manifests, Helm charts, or container orchestration definitions.
- Automated rollback or blue-green deployment mechanisms.
- Windows or macOS packaging.
