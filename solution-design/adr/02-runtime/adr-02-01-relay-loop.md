# ADR-02-01: Relay loop — blocking recv/send in dedicated thread

## Context

The relay loop is the performance-critical hot path. The current Rust implementation spawns a dedicated OS thread that blocks on `zmq_msg_recv`, then immediately calls `zmq_msg_send` — pure passthrough with no intermediate allocation or processing. This achieves ~750K msg/sec. The Zig conversion must preserve or exceed this throughput with zero per-message heap allocations.

See also: [PRD-001](../../prd/prd-001-high-throughput-message-relay.md), [03-pattern-map.md](../../03-pattern-map.md) patterns 1, 6, 10, 22

## Decision

The relay loop runs in a dedicated OS thread (`std.Thread.spawn`). It calls `c.zmq_msg_recv` and `c.zmq_msg_send` directly via `@cImport` with zero wrapper indirection. The message is a stack-allocated `c.zmq_msg_t` initialized with `zmq_msg_init` and cleaned up with `defer zmq_msg_close`.

The loop structure:
1. Init `zmq_msg_t` on the stack.
2. `zmq_msg_recv` (blocking) into the message.
3. `zmq_msg_send` the same message to the PUB socket.
4. If either call returns -1, log the error and break.
5. Check the running flag (`std.atomic.Value(bool).load(.monotonic)`).
6. Increment the message counter (`std.atomic.Value(usize).fetchAdd(1, .monotonic)`).
7. `zmq_msg_close` via defer, re-init for next iteration.

Socket options are set identically to the Rust code: `ZMQ_CONFLATE=0`, `ZMQ_IMMEDIATE=1`, `ZMQ_LINGER=0`, `ZMQ_RCVHWM=0`, `ZMQ_SNDHWM=0`, `ZMQ_XPUB_NODROP=1`.

### Alternatives considered

| Alternative | Why rejected |
|------------|-------------|
| Async I/O (Zig async, io_uring) | ZMQ's API is inherently blocking. `zmq_msg_recv` with `ZMQ_DONTWAIT` + polling adds complexity without throughput benefit for a single PULL→PUB pair. The blocking model already saturates 750K msg/sec. |
| Multiple relay threads | The current design uses one thread. ZMQ PULL socket distributes across multiple receivers but PUB fans out to all — adding relay threads would require careful message ordering. Out of scope per PRD-001. |
| `zmq_proxy` (built-in ZMQ relay) | `zmq_proxy` is a single C call that replaces the relay loop entirely. However, it provides no hook for the per-message counter (metrics), and its shutdown behaviour is different. Could be evaluated in the future but does not satisfy the metrics contract today. |

## Failure modes and mitigations

| Failure mode | Mitigation |
|-------------|-----------|
| `zmq_msg_recv` blocks indefinitely, preventing shutdown | The shutdown mechanism (see [ADR-02-02](adr-02-02-lifecycle-shutdown.md)) uses `zmq_ctx_shutdown` to interrupt blocking calls. When the context is shut down, `zmq_msg_recv` returns -1 with `errno=ETERM`, causing the loop to break. |
| `zmq_msg_send` fails on slow subscriber (ZMQ_XPUB_NODROP=1) | The loop logs the error and breaks. systemd restarts the binary. This matches the current Rust behaviour. |
| Per-message allocation in the hot path | `zmq_msg_t` is stack-allocated. `zmq_msg_init` + `zmq_msg_recv` use libzmq's internal buffer management (no Zig heap allocation). `std.fmt.bufPrintZ` for endpoint strings uses a stack buffer. No Zig allocator is used in the relay loop. |
| Atomic counter overhead on hot path | `fetchAdd(1, .monotonic)` is a single atomic instruction. On x86_64 this is `lock xadd` — ~20ns. Negligible relative to the `zmq_msg_recv`/`zmq_msg_send` cost (~1-2µs per message at 750K msg/sec). |

## Consequences

- The relay loop is ~30 lines of Zig calling libzmq C functions directly.
- No wrapper structs, no allocation, no indirection in the hot path.
- Throughput is bounded by libzmq's performance, not by the language layer.
- The per-message counter is an atomic increment — the only non-ZMQ operation in the loop.
- Interacts with [ADR-02-02](adr-02-02-lifecycle-shutdown.md) (shutdown), [ADR-02-03](adr-02-03-threading-coordination.md) (thread coordination), [ADR-02-04](adr-02-04-observability.md) (metrics counter).
