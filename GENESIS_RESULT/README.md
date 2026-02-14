# Solution Design — Lake arm64 Assembly Conversion

## Scope

This design covers the conversion of the Lake ZMQ message relay from Rust to hand-written AArch64 (arm64) assembly. Lake is a stateless, high-throughput ZMTP 3.0 PULL→PUB relay for the OpenBank platform, currently implemented as ~500 lines of Rust with unsafe FFI bindings to `libzmq`. The conversion targets Linux/arm64, preserving all external contracts (wire protocol, StatsD metrics, systemd integration, configuration interface) while eliminating the Rust compiler and standard library from the runtime.

## Key decisions

*To be populated after Phase 04 (Decisions).*

## How to read this doc set

| Document | Location | Purpose |
|----------|----------|---------|
| Architecture | [GENESIS_RESULT/architecture/](architecture/) | High-level architecture diagram and narrative |
| Prima materia | [GENESIS_SUPPORT/00-prima-materia.md](../GENESIS_SUPPORT/00-prima-materia.md) | Domain distillation and current-to-desired mapping |
| Constraints | [GENESIS_SUPPORT/01-constraints.md](../GENESIS_SUPPORT/01-constraints.md) | Target environment constraints, invariants, failure scenarios |
| Topology | [GENESIS_SUPPORT/02-topology.md](../GENESIS_SUPPORT/02-topology.md) | System topology, deployment units, component minimisation |
| Pattern map | [GENESIS_SUPPORT/03-pattern-map.md](../GENESIS_SUPPORT/03-pattern-map.md) | Pattern inventory, Rust→assembly mappings, unmappable patterns, external contracts |
| PRDs | [GENESIS_SUPPORT/prd/](../GENESIS_SUPPORT/prd/) | Product Requirements Documents |
| ADRs | [GENESIS_SUPPORT/adr/](../GENESIS_SUPPORT/adr/) | Architecture Decision Records |
| Migration | [GENESIS_SUPPORT/migration/](../GENESIS_SUPPORT/migration/) | Migration plan, coexistence, rollback |
