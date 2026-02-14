# Platform

| # | ADR | Summary |
|---|-----|---------|
| 01-01 | [Language — Zig](adr-01-01-language-zig.md) | Zig as implementation language; pinned version; direct C interop; eliminates all Rust dependencies |
| 01-02 | [Build system — Zig build](adr-01-02-build-system.md) | `build.zig` replaces Cargo; dual-arch cross-compilation; ReleaseFast optimization; `zig build` developer workflow |
| 01-03 | [ZMQ binding — direct C interop](adr-01-03-zmq-binding.md) | `@cImport("zmq.h")` replaces `zmq-sys` crate; no wrapper structs; `defer` for resource cleanup; inline error handling |
| 01-04 | [libzmq linking — dynamic](adr-01-04-libzmq-linking.md) | Dynamic linking against system `libzmq5`; preserves `.deb` dependency; static linking as future option |
