# ADR-02-04: Observability — hand-rolled StatsD client, /proc reader, custom log function

## Context

The relay must emit two StatsD metrics (`openbank.lake.message.relayed` count, `openbank.lake.memory.bytes` gauge) over UDP every second, and must produce structured log output to stdout for journald. The current Rust code uses the `statsd` crate (UDP client), `procfs` crate (/proc reader), `log` crate (facade), and `colored` crate (ANSI output). The topology eliminated the coloured output as a separate component (absorbed by logging). The conversion replaces all four crate dependencies with Zig standard library functionality and minimal hand-rolled code.

See also: [PRD-002](../../prd/prd-002-operational-observability.md), [PRD-005](../../prd/prd-005-conversion-contract-preservation.md), [03-pattern-map.md](../../03-pattern-map.md) patterns 13, 14, 20

## Decision

### StatsD client

A hand-rolled StatsD UDP client using `std.posix.socket` (AF_INET, SOCK_DGRAM). The client formats StatsD line protocol into a stack buffer via `std.fmt.bufPrint` and sends via `std.posix.sendto`. Fire-and-forget — UDP send errors are silently ignored (the relay must not stop due to metrics failure).

Metric format (StatsD line protocol):
- `openbank.lake.message.relayed:{count}|c`
- `openbank.lake.memory.bytes:{bytes}|g`

The client batches both metrics into a single UDP datagram (newline-separated) to match the Rust `statsd` crate's pipeline behaviour.

### Memory monitoring

Direct read of `/proc/self/statm` via `std.fs.openFileAbsolute` + `std.fs.File.reader`. Parse the second field (RSS in pages) and multiply by page size (`std.mem.page_size`). ~10 lines. Replaces the `procfs` crate and its transitive dependency tree (chrono, flate2, hex, etc.).

Platform dispatch: `if (comptime builtin.os.tag == .linux)` — on non-Linux (not a deployment target, but for compilation safety), return 0.

### Logging

A custom log function set via `pub const std_options` in the root source file. The function formats `YYYY-MM-DDTHH:MM:SSZ LVL [scope] message` to a stack buffer and writes to `std.io.getStdOut()`. ANSI colour codes are embedded as string literals (`"\x1b[32m"` for green, `"\x1b[31m"` for red, etc.) — no library needed.

Timestamp decomposition uses `std.time.timestamp()` (seconds since epoch) and manual year/month/day arithmetic matching the current Rust logger, or `std.time.epoch.EpochSeconds` if available in the pinned Zig version.

Log level is set at startup from `LAKE_LOG_LEVEL` environment variable. The `std.log` framework filters by level at compile time and runtime via `std_options.log_level`.

### Alternatives considered

| Alternative | Why rejected |
|------------|-------------|
| Third-party Zig StatsD library | None exists in the Zig ecosystem with sufficient maturity. The protocol is trivial (~30 lines). |
| Prometheus exposition endpoint instead of StatsD | Changes the metrics contract (invariant 2). Adds an HTTP server component (violates component budget). |
| `std.log` default formatter | Does not match the contracted log format. The format must be `YYYY-MM-DDTHH:MM:SSZ LVL [target] message`. |
| Use `syslog(3)` instead of stdout for logging | Changes the logging contract. The systemd unit captures stdout to journald. Switching to syslog would require unit file changes. |

## Failure modes and mitigations

| Failure mode | Mitigation |
|-------------|-----------|
| StatsD UDP send fails (daemon unreachable, network error) | Silently ignored. `sendto` returns an error; the metrics thread discards it and continues. The relay loop is unaffected. This matches the current Rust behaviour (failure scenario 3). |
| `/proc/self/statm` unreadable | Return 0 for memory bytes. Log a warning once. The relay continues operating. This cannot happen on a functional Linux system, but defensive coding costs nothing. |
| StatsD metric name or format drift from the Rust version | The metric names and StatsD encoding are string constants verified by the blackbox metrics test (`metrics.feature`). Any drift is caught by invariant 6 (tests pass unchanged). |
| Log format drift from the Rust version | The log format is verified by the blackbox configuration test (`configuration.feature`) which checks journalctl output. Any drift is caught by the test suite. |
| Stdout write blocks (pipe full, journald slow) | `std.io.getStdOut().write()` may block if the pipe buffer is full. This could stall the metrics thread (which does the logging in the current design). The relay thread does not log in the hot path — relay throughput is unaffected. A blocked log write delays metrics reporting, not message relay. |

## Consequences

- Four Rust crate dependencies (`statsd`, `procfs`, `log`, `colored`) are eliminated. Zero Zig third-party dependencies for observability.
- The metrics thread is ~60 lines of Zig: UDP socket setup, `/proc` read, StatsD format, send.
- The log function is ~40 lines of Zig: timestamp, level color, format, stdout write.
- All observability code is internal to the binary — no external components added.
- Interacts with [ADR-02-03](adr-02-03-threading-coordination.md) (metrics thread reads the atomic counter), [ADR-02-02](adr-02-02-lifecycle-shutdown.md) (metrics thread exits on running flag).
