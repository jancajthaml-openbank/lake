# PRD-003: Daemon lifecycle and host integration

## Problem / requirement

The relay runs as a long-lived daemon on Linux hosts managed by systemd. It must start quickly, signal readiness to systemd, run indefinitely, and shut down cleanly when told to stop. If the binary does not send the readiness signal within 1 second, systemd considers the start failed. If it does not exit within 3 seconds of receiving SIGTERM, systemd forcibly kills it, producing unclean shutdown evidence in the system journal that the operator must investigate.

The user's stated intent includes making the system "more reliable." The current implementation uses a workaround for shutdown: it sends a dummy message to its own PULL socket to unblock the blocking receive call, and worker threads raise SIGTERM to coordinate exit with the main thread. This is fragile — if the self-connect fails or the signal is lost, the thread join deadlocks and systemd SIGKILL follows. The conversion is an opportunity to make shutdown more deterministic.

Configuration is loaded from environment variables at startup, with defaults for all parameters. The systemd unit uses an EnvironmentFile at a well-known path. A filesystem watcher unit restarts the service when the configuration file changes. The binary itself does not hot-reload configuration.

## Success criteria

- The binary must send `READY=1` to the systemd notification socket (`$NOTIFY_SOCKET`) within 1 second of process start.
- The binary must send `STOPPING=1` to the systemd notification socket when shutdown begins.
- The binary must exit cleanly (exit code 0) within 3 seconds of receiving SIGTERM.
- All worker threads must be joined and all resources (sockets, contexts, file descriptors) must be released before the process exits.
- The binary must read configuration from environment variables `LAKE_PORT_PULL` (default 5562), `LAKE_PORT_PUB` (default 5561), `LAKE_LOG_LEVEL` (default INFO), and `LAKE_STATSD_ENDPOINT` (default `127.0.0.1:8125`).
- The binary must log startup and shutdown events (at minimum: "Program starting", "Program stopping", "Relay started", "Relay stopped").
- The shutdown mechanism must not depend on the relay's own PULL socket being connectable — if the self-connect pattern is retained, its failure must not prevent shutdown.

## Out of scope

- Configuration hot-reload within the binary (systemd watcher handles this by restarting).
- Watchdog integration (`sd_notify WATCHDOG=1`).
- Multi-instance or socket-activated startup.
- Non-Linux platforms (macOS, Windows).
