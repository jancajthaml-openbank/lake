# Solution Design — Lake

## Scope

Full conversion of the Lake message relay from Rust to Zig, preserving all 13 external contracts (ZMQ PULL/PUB on ZMTP 3.0, StatsD metrics, systemd sd_notify, environment variable configuration, Debian packaging, Docker images) while improving reliability (deterministic shutdown via `zmq_ctx_shutdown`, eliminating unsafe FFI indirection) and performance (zero per-message heap allocation in the relay loop, stack-allocated atomics replacing `Arc`). The relay is a stateless, high-throughput (~750K msg/sec) ZMQ message forwarder for the OpenBank platform, deployed as a single binary on commodity Linux hosts.

## Key decisions

- **[ADR-01-01 — Language: Zig](adr/01-platform/adr-01-01-language-zig.md)**: Zig (pinned 0.13.x) replaces Rust. Direct C interop via `@cImport` eliminates 6 Rust crate dependencies. All 31 `unsafe` blocks become natural C calls.
- **[ADR-01-03 — ZMQ binding: direct C interop](adr/01-platform/adr-01-03-zmq-binding.md)**: `@cImport("zmq.h")` — no wrapper structs, no error module, no message abstraction. The topology eliminated 3 Rust abstractions as unnecessary in Zig.
- **[ADR-02-01 — Relay loop: zero-alloc hot path](adr/02-runtime/adr-02-01-relay-loop.md)**: Blocking `zmq_msg_recv` → `zmq_msg_send` in dedicated thread. Stack-allocated `zmq_msg_t`. Zero Zig allocator usage. Throughput bounded by libzmq, not the language layer.
- **[ADR-02-02 — Shutdown: zmq_ctx_shutdown](adr/02-runtime/adr-02-02-lifecycle-shutdown.md)**: Replaces the fragile self-PUSH trick with libzmq's official context shutdown mechanism. Deterministic: set flag → shutdown context → join threads → exit. Completes in ~1-2 seconds.
- **[ADR-02-03 — Threading: stack atomics](adr/02-runtime/adr-02-03-threading-coordination.md)**: `std.atomic.Value` on main's stack with pointer sharing. No `Arc`, no heap allocation, no refcounting. Strict linear lifecycle guarantees safety.
- **[ADR-01-04 — libzmq: dynamic linking](adr/01-platform/adr-01-04-libzmq-linking.md)**: Dynamic linking against system `libzmq5` preserves `.deb` dependency contract. Static linking available as build flag for future container deployments.

## How to read this doc set

| Document | Purpose |
|----------|---------|
| [00-prima-materia.md](00-prima-materia.md) | Domain distillation and current-to-desired mapping with ADR links |
| [01-constraints.md](01-constraints.md) | Target environment constraints, invariants, failure scenarios |
| [02-topology.md](02-topology.md) | System topology, deployment units, communication patterns, component minimisation, reduction pass |
| [03-pattern-map.md](03-pattern-map.md) | Rust→Zig pattern inventory, mappings, unmappable patterns, external contracts |
| [prd/](prd/) | Product Decision Records (6 domain-problem PRDs) |
| [adr/](adr/) | Architecture Decision Records (10 ADRs across 3 areas) |
| [architecture/](architecture/) | High-level architecture diagram, narratives, traffic summary |
| [migration/](migration/) | Migration plan, coexistence, rollback (Phase 05.5) |
