# Runtime

| # | ADR | Summary |
|---|-----|---------|
| 02-01 | [Relay loop — blocking recv/send](adr-02-01-relay-loop.md) | Dedicated OS thread; direct `zmq_msg_recv`→`zmq_msg_send` via C calls; zero per-message heap allocation; stack-allocated `zmq_msg_t` |
| 02-02 | [Lifecycle and shutdown — zmq_ctx_shutdown](adr-02-02-lifecycle-shutdown.md) | `zmq_ctx_shutdown` replaces self-PUSH trick; deterministic shutdown sequence; sd_notify integration; <3s shutdown |
| 02-03 | [Threading — stack atomics with pointer sharing](adr-02-03-threading-coordination.md) | `std.atomic.Value` on main's stack; pointer sharing to threads; no Arc, no heap allocation; explicit context structs |
| 02-04 | [Observability — StatsD, /proc, logging](adr-02-04-observability.md) | Hand-rolled StatsD UDP client; direct `/proc/self/statm` read; custom `std.log` function; ANSI colour inline; zero dependencies |
