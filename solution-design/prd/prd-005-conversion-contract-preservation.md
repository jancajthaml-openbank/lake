# PRD-005: Conversion contract preservation

## Problem / requirement

The relay is one component in the OpenBank platform. Other services connect to it via ZMQ, monitoring systems consume its StatsD metrics, systemd manages its lifecycle, and operators install it via `.deb` packages. Each of these integrations constitutes an external contract. If any contract changes during the language conversion, the downstream integration breaks — and the breakage may be silent (wrong metric names, subtly different wire behaviour, different exit codes).

The conversion from Rust to Zig must be invisible to every external consumer. A ZMQ client that connects to the relay must not be able to distinguish the Zig binary from the Rust binary. A StatsD dashboard must not need reconfiguration. A systemd unit file must not need modification. A deployment script that installs the `.deb` must not need updating. The blackbox test suite — which exercises all of these contracts — must pass without modification.

This is the verification gate: if the existing tests pass unchanged against the Zig binary, the contracts are preserved. If any test requires modification, a contract was broken.

## Success criteria

- All 6 existing blackbox test features (install, configuration, management, metrics, relay, uninstall) must pass against the Zig binary without any test code modifications.
- The ZMTP 3.0 wire protocol behaviour must be identical — any ZMQ PUSH client connecting to the PULL port, and any ZMQ SUB client connecting to the PUB port, must work identically with no client-side changes.
- The StatsD metric names (`openbank.lake.message.relayed`, `openbank.lake.memory.bytes`), types (count, gauge), prefix (`openbank.lake`), and UDP encoding must be identical.
- The systemd notification protocol (READY=1, STOPPING=1) must be identical.
- The environment variable names, defaults, and parsing behaviour must be identical.
- The binary name (`lake`), install path (`/usr/bin/lake`), and process signal behaviour (clean exit on SIGTERM) must be identical.
- The log output format (`YYYY-MM-DDTHH:MM:SSZ LVL [target] message`) must be identical.

## Out of scope

- Extending contracts (new metrics, new env vars, new socket types) — that is a future concern, not a conversion concern.
- Backward compatibility with Rust source-level APIs (there are no library consumers).
- Contract documentation beyond what the test suite already covers.
