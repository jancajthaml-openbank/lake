# ADR-02-02: Process lifecycle and shutdown — zmq_ctx_shutdown + sigwait

## Context

The binary must start within 1 second (sd_notify `READY=1`) and shut down within 3 seconds of SIGTERM (systemd `TimeoutStopSec=3`). The current Rust implementation has a fragile shutdown path: the relay's `Drop` impl sends a dummy PUSH message to its own PULL socket to unblock the blocking `zmq_msg_recv`, and both worker threads raise `SIGTERM` via `libc::raise` to unblock the main thread's `sigwait`. If any of these coordination steps fails, the thread join deadlocks and systemd SIGKILLs the process (failure scenario 10). The user's intent includes "more reliable" — this is an opportunity to improve shutdown determinism.

See also: [PRD-003](../../prd/prd-003-daemon-lifecycle.md), [03-pattern-map.md](../../03-pattern-map.md) patterns 3, 4, 7, 11, 12

## Decision

### Startup sequence

1. Load configuration from environment variables.
2. Initialize the custom log function.
3. Send `sd_notify READY=1` to `$NOTIFY_SOCKET` via Unix datagram (using `std.posix.socket`, `std.posix.sendto`).
4. Log "Program starting".
5. Create ZMQ context (`zmq_ctx_new`).
6. Spawn metrics thread.
7. Spawn relay thread.
8. Log "Relay started".
9. Block on `sigwait` for SIGTERM.

### Shutdown sequence (after SIGTERM received)

1. Set the running flag to `false` (`running.store(false, .monotonic)`).
2. Call `c.zmq_ctx_shutdown(ctx)` — this causes any blocking `zmq_msg_recv` in the relay thread to return -1 with `errno=ETERM`, breaking the relay loop without a self-PUSH trick.
3. Join the relay thread (`relay_thread.join()`).
4. Join the metrics thread (the metrics thread checks the running flag each second and exits when false).
5. Call `c.zmq_ctx_term(ctx)` — cleans up the context after all sockets are closed.
6. Send `sd_notify STOPPING=1`.
7. Log "Program stopping".
8. Exit with code 0.

### Why `zmq_ctx_shutdown` instead of self-PUSH

The Rust code's self-PUSH pattern creates a temporary ZMQ context and PUSH socket, connects to localhost, and sends an empty message to unblock `zmq_msg_recv`. This fails if: the PULL port is unavailable, the connect fails, or the context creation fails. `zmq_ctx_shutdown` is libzmq's official mechanism for interrupting blocking calls — it signals all sockets in the context to return `ETERM`. It requires no network operation, no temporary socket, and no coordination. It is strictly more reliable.

### Alternatives considered

| Alternative | Why rejected |
|------------|-------------|
| Keep the self-PUSH shutdown trick | Fragile — depends on network connectivity to localhost. `zmq_ctx_shutdown` is simpler and more reliable. |
| Use `zmq_msg_recv` with `ZMQ_DONTWAIT` + polling | Adds per-iteration overhead (poll syscall) to the hot path. The blocking model with `zmq_ctx_shutdown` for interruption is simpler and faster. |
| Use `pthread_cancel` to terminate the relay thread | Unsafe — leaves ZMQ resources in an undefined state. |
| Use `signalfd` instead of `sigwait` | Functionally equivalent for our use case (single signal, main thread only). `sigwait` is simpler and more portable. |

## Failure modes and mitigations

| Failure mode | Mitigation |
|-------------|-----------|
| `zmq_ctx_shutdown` does not unblock `zmq_msg_recv` (libzmq bug) | Defensive timeout: if `relay_thread.join()` has not returned within 2 seconds, log a warning. systemd will SIGKILL after 3 seconds. This is a last resort — `zmq_ctx_shutdown` is reliable in libzmq 4.3.x. |
| `sd_notify` socket not available (`$NOTIFY_SOCKET` unset) | Check for the env var before sending. If unset, skip silently. This is safe: when running outside systemd (e.g., manually for debugging), no notification is needed. |
| Metrics thread does not exit within shutdown window | The metrics thread sleeps for 1 second between iterations. Worst case, it takes 1 second to notice the running flag is false. Combined with relay thread join, shutdown completes in ~1-2 seconds — well within the 3-second budget. |
| SIGTERM arrives before startup completes | The `sigwait` blocks until SIGTERM is received. If startup fails (e.g., socket bind error), the process exits before reaching `sigwait`. If SIGTERM arrives during startup, `sigwait` returns immediately and shutdown proceeds normally. |

## Consequences

- The self-PUSH shutdown pattern is eliminated — no temporary context, no temporary socket, no loopback connection.
- Shutdown is deterministic: `zmq_ctx_shutdown` → thread join → `zmq_ctx_term` → exit.
- The `kill_port` field in the Rust `Relay` struct has no equivalent — the relay thread needs no reference to the PULL port for shutdown.
- Worker threads no longer call `libc::raise(SIGTERM)` — the main thread's `sigwait` handles SIGTERM directly; worker threads exit via the running flag and `ETERM` from `zmq_ctx_shutdown`.
- Interacts with [ADR-02-01](adr-02-01-relay-loop.md) (relay loop breaks on ETERM), [ADR-02-03](adr-02-03-threading-coordination.md) (thread join order).
