# PRD-001: High-throughput message relay

## Problem / requirement

The OpenBank platform requires a message relay that receives messages from upstream services and fans them out to downstream subscribers at high throughput. The relay is the communication backbone — if it is slow, every service that depends on it is slow; if it drops messages, downstream state diverges.

The current implementation achieves approximately 750,000 messages per second on commodity hardware (1 CPU, 2 GB RAM). The conversion must preserve or exceed this throughput. The relay must remain stateless — messages pass through without persistence, filtering, or transformation. The relay loop is the hot path: every CPU cycle spent on overhead in this loop directly reduces the platform's messaging capacity.

The user's stated intent includes making the system "more performant." This means the conversion is not merely a language swap — it is an opportunity to reduce per-message overhead (allocation, indirection, system call wrapping) in the relay loop.

## Success criteria

- The relay must receive messages on a configurable ingest port and fan them out on a configurable publish port using the ZMTP 3.0 wire protocol.
- Throughput must be at least 750,000 messages per second on equivalent hardware (1 CPU, 2 GB RAM, amd64), as verified by the existing performance test suite.
- The relay loop must introduce zero heap allocations per message in the steady state (recv → send path).
- Message ordering must be preserved: messages received in order A, B, C must be published in order A, B, C.
- The relay must not block, crash, or leak resources when a subscriber is slow — backpressure behaviour must match the current implementation (error on send to slow subscriber rather than silent drop, per the existing socket option configuration).
- The relay must continue operating when the metrics subsystem is unavailable or degraded.

## Out of scope

- Message persistence or replay.
- Message filtering, transformation, or routing by content.
- Multi-port or multi-topic relay (the relay handles one PULL→PUB pair).
- Encryption or authentication of the messaging channel (the relay uses the NULL security mechanism).
- Throughput beyond a single CPU core (the relay runs one relay thread by design).
