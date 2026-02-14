# ADR-02-03: Threading and cross-thread coordination — stack atomics with pointer sharing

## Context

The relay uses three OS threads: main (signal wait + lifecycle), relay (PULL→PUB loop), and metrics (StatsD reporter). These threads share two pieces of mutable state: a boolean running flag (shutdown signal) and a counter (messages relayed). The current Rust code uses `Arc<AtomicBool>` and `Arc<AtomicUsize>` — heap-allocated, reference-counted atomics. Zig has no `Arc`; the pattern map (patterns 4, 5, 6, 7) identifies this as a structural translation.

See also: [PRD-001](../../prd/prd-001-high-throughput-message-relay.md), [PRD-003](../../prd/prd-003-daemon-lifecycle.md), [03-pattern-map.md](../../03-pattern-map.md) patterns 4, 5, 6, 7

## Decision

Shared state is allocated on the main thread's stack as `std.atomic.Value` instances. Pointers to these atomics are passed to spawned threads via explicit context structs.

```
main() {
    var running = std.atomic.Value(bool).init(true);
    var msg_count = std.atomic.Value(usize).init(0);

    // spawn relay thread, passing &running and &msg_count
    // spawn metrics thread, passing &running and &msg_count
    // sigwait...
    // running.store(false, .monotonic)
    // join threads
}
```

Thread context structs are defined for each thread:

- `RelayContext`: ZMQ context pointer, PULL port, PUB port, `*std.atomic.Value(bool)` (running), `*std.atomic.Value(usize)` (msg_count).
- `MetricsContext`: StatsD endpoint string, `*std.atomic.Value(bool)` (running), `*std.atomic.Value(usize)` (msg_count).

Threads are spawned with `std.Thread.spawn(.{}, relayThreadFn, .{&relay_ctx})` and joined with `relay_thread.join()` during shutdown.

### Why stack allocation instead of heap (Arc equivalent)

The program has a strict linear lifecycle: main creates shared state → spawns threads → waits for signal → sets running to false → joins all threads → exits. The shared state's lifetime is the main function's scope. Threads are always joined before main returns. Therefore, the shared state always outlives all threads — no reference counting needed.

This eliminates: 2 heap allocations (Arc boxes), 2 atomic reference counts, and their associated cache-line contention. The atomics live on main's stack, which is hot in L1 cache.

### Alternatives considered

| Alternative | Why rejected |
|------------|-------------|
| Global mutable state (`var running: ...` at module level) | Zig supports this, but it couples the lifetime to the process rather than to the function scope. Harder to reason about in tests. |
| Heap-allocated atomics (Zig allocator + manual free) | Unnecessary — stack allocation is sufficient given the linear lifecycle. Adds allocation failure handling for no benefit. |
| Channel / message passing between threads | Overkill for a single boolean flag and a counter. Channels introduce allocation and synchronisation overhead in the hot path. |
| Mutex-protected shared struct | Atomics are lock-free and sufficient for a bool and a usize. A mutex would add unnecessary contention. |

## Failure modes and mitigations

| Failure mode | Mitigation |
|-------------|-----------|
| Thread accesses atomic after main's stack frame is destroyed | Impossible if threads are joined before main returns. The shutdown sequence in [ADR-02-02](adr-02-02-lifecycle-shutdown.md) guarantees this: join relay thread → join metrics thread → main returns. |
| Thread spawn fails (OS resource exhaustion) | `std.Thread.spawn` returns an error. Handle it by logging and exiting — the binary cannot function without its relay thread. systemd will restart. |
| Atomic ordering too weak (`.monotonic`) causes stale reads | For the running flag: the worst case is one extra loop iteration (1µs at 750K msg/sec) before the relay thread sees the shutdown signal. Acceptable — well within the 3-second shutdown window. For the message counter: monotonic is sufficient because the counter is read by a single thread (metrics) once per second. No consistency guarantee needed between reads. |
| Context struct contains dangling pointer after stack reallocation | Zig does not reallocate stacks. `std.Thread.spawn` captures the context by value at spawn time. The context struct contains pointers to main's stack — these are stable because main blocks on `sigwait` and the stack frame persists until after threads are joined. |

## Consequences

- No heap allocations for shared state.
- No reference counting overhead.
- Thread-safety relies on the disciplined lifecycle (spawn → sigwait → join) rather than on compiler-enforced ownership. Documented in code comments.
- The `Arc` pattern from Rust is replaced entirely — this is one of the three unmappable patterns from the pattern map, resolved by architectural simplification.
- Interacts with [ADR-02-01](adr-02-01-relay-loop.md) (relay thread reads running flag and increments counter), [ADR-02-02](adr-02-02-lifecycle-shutdown.md) (main sets running to false, joins threads), [ADR-02-04](adr-02-04-observability.md) (metrics thread reads and resets counter).
