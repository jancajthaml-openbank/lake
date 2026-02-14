# PRD-002: Operational observability

## Problem / requirement

The operator needs to know whether the relay is alive, how many messages it is processing, and how much memory it is consuming — without attaching a debugger, reading logs in real time, or instrumenting the binary externally. The current implementation reports two metrics to a StatsD daemon over UDP on a one-second interval: message throughput count and process RSS memory usage. Downstream dashboards and alerting depend on these specific metric names and their StatsD encoding.

The operator also needs structured log output captured by journald for post-incident diagnosis. The log format includes ISO-8601 UTC timestamps, colored severity levels, and a module target identifier. Downstream log aggregation or parsing tools may depend on this format.

The observability subsystem must not interfere with the relay's primary function. If the StatsD daemon is unreachable, the relay must continue relaying messages. If logging to stdout blocks or is slow, it must not stall the relay loop.

## Success criteria

- The binary must emit a StatsD counter metric named `openbank.lake.message.relayed` reflecting the number of messages relayed since the last reporting interval.
- The binary must emit a StatsD gauge metric named `openbank.lake.memory.bytes` reflecting the process's resident memory size in bytes.
- Metrics must be sent to the endpoint specified by the `LAKE_STATSD_ENDPOINT` environment variable (default `127.0.0.1:8125`) over UDP.
- Metrics reporting must not crash, block, or cause the relay to stop if the StatsD endpoint is unreachable.
- Log output must be written to stdout in the format `YYYY-MM-DDTHH:MM:SSZ LVL [target] message` for capture by journald.
- Log level must be configurable via the `LAKE_LOG_LEVEL` environment variable (DEBUG, INFO, WARN, ERROR) with a default of INFO.
- The metrics and logging subsystems must run independently of the relay loop — a failure in observability must not degrade message throughput.

## Out of scope

- Prometheus or OpenTelemetry metrics export.
- Log file rotation or log-to-file output (journald handles this).
- Distributed tracing or request-level correlation IDs.
- Custom metric names or user-defined metric dimensions.
