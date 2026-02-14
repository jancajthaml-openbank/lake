# Solution Design — Lake

## Scope

Full conversion of the Lake message relay from Rust to Zig, preserving all external contracts (ZMQ PULL/PUB on ZMTP 3.0, StatsD metrics, systemd sd_notify, environment variable configuration, Debian packaging, Docker images) while improving reliability (eliminating unsafe FFI indirection, strengthening error handling) and performance (reducing allocation overhead, leveraging Zig's direct C interop and comptime capabilities). The relay is a stateless, high-throughput (~750K msg/sec) ZMQ message forwarder for the OpenBank platform.

## Key decisions

*To be populated during Phase 04 (Decisions).*

## How to read this doc set

| Document | Purpose |
|----------|---------|
| [00-prima-materia.md](00-prima-materia.md) | Domain distillation and current-to-desired mapping |
| [01-constraints.md](01-constraints.md) | Target environment constraints, invariants, failure scenarios |
| [02-topology.md](02-topology.md) | System topology, deployment units, component minimisation |
| [03-pattern-map.md](03-pattern-map.md) | Pattern inventory, mappings, unmappable patterns, contracts |
| [prd/](prd/) | Product Decision Records |
| [adr/](adr/) | Architecture Decision Records |
| [architecture/](architecture/) | High-level architecture |
| [migration/](migration/) | Migration plan, coexistence, rollback |
