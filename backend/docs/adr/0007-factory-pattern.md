# 7. Use a factory to build a RateLimiter from configuration

## Status

Accepted

## Context

There are five algorithms x two backends (see [0001](0001-five-rate-limiting-algorithms.md)
and [0002](0002-memory-and-redis-backends.md)) — ten concrete
constructors — but callers (`main.go`, tests, benchmarks) just want
"give me a working `RateLimiter` for this config." Something has to own
the mapping from `RateLimitConfig` to the right constructor call, with
the right validation, for the right backend.

## Decision

Centralize that mapping in `internal/limiter/factory.Create(cfg) (RateLimiter, io.Closer, error)`.

## Why

Without it, that ten-way branch (which algorithm, which backend, which
validation rules apply, which Redis client to construct) would either
be duplicated everywhere a limiter is needed, or `main.go` would need
to know the internal details of all ten implementations. Neither is
appealing: the first duplicates non-trivial logic, the second turns the
composition root into a place that understands every algorithm's
Redis-vs-memory constructor signature and its own capacity/limit field
name (`Limit` for window-based algorithms, `Capacity` for bucket-based
ones).

The factory also owns per-algorithm config validation (a fixed window
with `Limit: 0` is rejected before a limiter is ever constructed) and —
important for the Redis case — resource ownership: it returns an
`io.Closer` alongside the limiter, a no-op for memory storage and the
real Redis client for Redis storage. Callers always call `Close()`
without needing a type assertion or knowing which backend they got.

## Consequences

- Adding a sixth algorithm means adding one function to
  `factory.go` and one branch in `Create`'s switch — not touching every
  caller.
- The factory is the one place that has to know about every concrete
  package (`fixedwindow`, `slidinglog`, ..., `adapter`, `redisstore`) —
  by design; everything else only knows about the `RateLimiter`
  interface.
- Tests that need a real limiter (integration tests, the distributed
  correctness test, benchmarks) go through the factory rather than
  constructing algorithms directly, so they exercise the same
  construction path production code does.
