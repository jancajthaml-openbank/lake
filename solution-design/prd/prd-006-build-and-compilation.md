# PRD-006: Build and compilation infrastructure

## Problem / requirement

The relay must be buildable, testable, and releasable using the Zig toolchain. The current build infrastructure (Cargo, Rust compiler, clang-13 linker, custom Docker images for CI) must be replaced with Zig equivalents. The build system must produce stripped, optimized binaries for both amd64 and arm64 Linux, and must integrate with the existing CI pipelines (CircleCI, GitHub Actions) and packaging scripts (Debian, Docker).

Zig is a pre-1.0 language. Breaking changes between compiler versions are expected. The build system must pin a specific Zig compiler version so that CI builds are reproducible and a Zig release does not silently break the build. The project must not depend on external Zig package registries at build time for its core functionality — the only external C dependency is libzmq.

The development team must be able to build, test, and package the relay with a minimal local setup: the Zig compiler and libzmq development headers. The build must support both dynamic linking (against system libzmq5) and the option for static linking (bundling libzmq into the binary to reduce runtime dependencies).

## Success criteria

- A `build.zig` file must define the build for the relay binary, supporting both `x86_64-linux` and `aarch64-linux` targets.
- The build must produce a single statically-linked or dynamically-linked binary (the linking strategy must be an explicit build option, not accidental).
- The build must produce a stripped, release-optimized binary comparable in size to the current Rust binary (a few MB or smaller).
- The Zig compiler version must be pinned in the build definition and CI configuration.
- The CI pipeline must build, test, and package the binary for both architectures.
- The build must integrate with the existing `dev/lifecycle/` scripts or replace them with equivalent Zig build steps.
- A developer must be able to build the binary locally with `zig build` after installing the Zig compiler and libzmq development headers.
- The build must support running Zig tests (`zig build test`) for any unit-level tests written in the Zig source.

## Out of scope

- IDE integration or editor plugins for Zig.
- Zig package registry publishing.
- Automated Zig compiler version upgrades.
- Build support for non-Linux targets (macOS, Windows) beyond cross-compilation from Linux.
