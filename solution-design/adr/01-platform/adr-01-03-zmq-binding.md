# ADR-01-03: ZMQ binding — direct C interop via @cImport

## Context

The relay's core function depends on the libzmq C library for ZMTP 3.0 messaging. The current Rust code uses the `zmq-sys` crate (raw FFI bindings) with 31 `unsafe` blocks across 6 files, plus custom wrapper structs (`Context`, `Socket`, `Message`) and a 27-variant error enum to provide safe Rust abstractions over the C API. The topology (Phase 03.5) eliminated these wrappers as unnecessary in Zig — direct C calls with `defer` cleanup replace the wrapper layer.

See also: [PRD-001](../../prd/prd-001-high-throughput-message-relay.md), [03-pattern-map.md](../../03-pattern-map.md) patterns 1, 3, 10, 15, 16, 22

## Decision

We use Zig's `@cImport` to import libzmq headers directly. All ZMQ functions (`zmq_ctx_new`, `zmq_socket`, `zmq_bind`, `zmq_msg_recv`, `zmq_msg_send`, etc.) are called as C functions with no wrapper layer.

```zig
const c = @cImport({
    @cInclude("zmq.h");
});
```

Resource cleanup uses `defer` at each creation site:
- Context: `defer { _ = c.zmq_ctx_term(ctx); }`
- Socket: `defer { _ = c.zmq_close(sock); }`
- Message: `defer { _ = c.zmq_msg_close(&msg); }`

Error handling is inline: check return value, log via `c.zmq_strerror(c.zmq_errno())`, return a Zig error from a small error set (`error{SocketError, BindError, ContextError}`).

No wrapper structs. No separate error module. No separate message module. The topology eliminated all three as absorbed by the relay code.

### Alternatives considered

| Alternative | Why rejected |
|------------|-------------|
| Zig wrapper structs mirroring Rust's `Context`/`Socket`/`Message` | Adds indirection without safety benefit — Zig has no `unsafe` boundary to abstract over. `defer` achieves RAII without a struct. |
| Third-party Zig ZMQ binding library | Adds an external dependency; the binding is ~10 lines of `@cImport`. No value added. |
| Re-implement ZMTP protocol in Zig (no libzmq) | Massive effort with high risk of protocol incompatibility. libzmq is battle-tested. Violates the reliability intent. |

## Failure modes and mitigations

| Failure mode | Mitigation |
|-------------|-----------|
| `@cImport` fails to parse `zmq.h` | libzmq's C header is well-formed and stable across 4.3.x. Zig's `@cImport` handles standard C headers reliably. Pin libzmq version in `.deb` dependency. |
| Missing `defer` causes ZMQ resource leak (socket, context, message not closed) | Code review discipline: every `zmq_ctx_new` / `zmq_socket` / `zmq_msg_init` must have a corresponding `defer` on the next line. Zig's `GeneralPurposeAllocator` leak detection covers heap allocations; fd leaks are caught by monitoring `/proc/self/fd` count in extended tests. |
| ZMQ C function returns error but caller ignores return value | Zig does not allow ignoring return values by default (`error: unused function return value`). For functions returning `c_int`, wrap in a helper that checks `== -1` and converts to a Zig error. For fire-and-forget calls (e.g., `zmq_close` in `defer`), explicitly discard with `_ = c.zmq_close(...)`. |
| Semantic mismatch between C null pointer and Zig optional | `zmq_ctx_new()` and `zmq_socket()` return `?*anyopaque` through `@cImport`. Zig forces a null check via `orelse` — this is safer than the Rust code which checked `.is_null()` manually. |

## Consequences

- The Rust modules `socket.rs`, `message.rs`, and `error.rs` have no Zig equivalents. Their functionality is absorbed inline.
- The `zmq-sys` crate dependency is eliminated.
- The binary links against `libzmq5.so` at runtime (dynamic) or embeds it at build time (static) — see [ADR-01-04](adr-01-04-libzmq-linking.md).
- ZMQ socket options (`ZMQ_CONFLATE`, `ZMQ_IMMEDIATE`, `ZMQ_LINGER`, `ZMQ_RCVHWM`, `ZMQ_SNDHWM`, `ZMQ_XPUB_NODROP`) are set via `c.zmq_setsockopt()` with identical values to the Rust code.
- The relay loop calls `c.zmq_msg_recv` and `c.zmq_msg_send` directly — zero indirection in the hot path.
