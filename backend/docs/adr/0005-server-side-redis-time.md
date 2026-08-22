# 5. Use Redis's server-side TIME instead of the client's wall clock

## Status

Accepted

## Context

Sliding log, sliding counter, token bucket, and leaky bucket all need
"the current time" to decide whether enough time has passed for a
window to roll over, a token to refill, or a queue slot to leak. That
time could come from the Go client's `time.Now()` or from Redis's own
`TIME` command.

## Decision

Every Redis-backed algorithm's Lua script calls `redis.call("TIME")`
for its notion of "now," rather than receiving a timestamp computed by
the calling Go process.

## Why

The whole point of the Redis backend (see [0002](0002-memory-and-redis-backends.md))
is that multiple app instances share one source of truth. If each
instance instead sent its own `time.Now()` into the script, they'd
each be asserting a *different* "now" for the same shared state —
and those clocks are not guaranteed to agree. Container clocks drift.
NTP sync isn't instantaneous or perfectly accurate. A few hundred
milliseconds of skew between two instances is enough to make one of
them see a window as already rolled over while the other still sees it
active, for the same key, at effectively the same real moment.

Asking Redis for the time — since Redis itself is the single shared
component every instance talks to — means every instance's script
execution agrees on "now," regardless of clock skew on the machines
running the app.

## Consequences

- Every timing calculation happens inside the Lua script, in
  milliseconds derived from `TIME`, rather than being computed in Go
  and passed in as an argument (window size and rates are passed in;
  the current moment is not).
- The memory-backed implementations still use `time.Now()` directly,
  since there's only one process and clock skew across processes isn't
  a concern for them — this is a Redis-specific decision, not a
  project-wide one.
