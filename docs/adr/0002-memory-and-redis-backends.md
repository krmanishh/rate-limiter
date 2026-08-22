# 2. Support both memory and Redis backends

## Status

Accepted

## Context

A rate limiter's state (counts, timestamps, tokens, queue sizes) has to
live somewhere. It can live in the process's own memory, or in a
shared store multiple processes can see.

## Decision

Support both, behind the same `RateLimiter` interface, selectable via
`RATE_LIMIT_STORAGE=memory|redis`.

## Why

These solve different problems and neither one is strictly better:

- **Memory** is fast (tens of nanoseconds per decision — see the
  README's Performance section) and needs no external dependency. But
  each process has its own state: run three instances behind a load
  balancer with a memory-backed limit of 5, and the *effective* limit
  becomes 15, not 5, because no instance knows what the others have
  allowed.
- **Redis** fixes exactly that: state lives in one shared place, so the
  limit is enforced across every instance talking to it (see
  [`docs/adr/0004`](0004-lua-scripts-for-atomicity.md) for how that's
  made atomic, and the README's distributed correctness test for proof
  it actually works). The cost is real: roughly 2,000-3,000x the
  latency of memory, per the benchmarks — a real network round trip and
  Lua script execution replace a mutex and a map lookup.

The right choice depends entirely on deployment shape. A single
instance has no reason to pay Redis's cost. Multiple instances sharing
a limit have no way to get correct behavior without something like
Redis. Making this a runtime choice rather than a fixed decision means
the same code serves both cases.

## Consequences

- Every algorithm needs two implementations (see
  [0001](0001-five-rate-limiting-algorithms.md)), not one.
- The `RateLimiter` interface has to be storage-agnostic — it can't
  leak anything Redis-specific (a context, an error) into its shape.
  That constraint is what led to the adapter pattern
  ([0008](0008-redis-adapter.md)).
- Memory-backed limiters can never error (there's no I/O to fail), but
  Redis-backed ones can — see [0006](0006-fail-open-vs-fail-closed.md)
  for how that's handled.
