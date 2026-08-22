# 1. Implement five rate limiting algorithms, not one

## Status

Accepted

## Context

A rate limiter needs exactly one algorithm to be useful in production.
This project implements five: fixed window, sliding log, sliding
counter, token bucket, and leaky bucket.

## Decision

Build all five behind the same `RateLimiter` interface, selectable at
runtime via `RATE_LIMIT_ALGORITHM`, rather than picking one and
building it well.

## Why

Each algorithm makes a genuinely different trade-off, and the
differences are easiest to understand by having working, comparable
implementations side by side rather than reading about them:

- **Fixed window** is simple and cheap but allows a burst of up to 2x
  the limit at window boundaries (a client can spend its whole quota at
  the end of one window and again at the start of the next).
- **Sliding log** is exact — no boundary bursting — but its memory cost
  scales with request volume, since it keeps a timestamp per request.
- **Sliding counter** approximates a sliding window with two fixed
  windows and a weighted average, trading exactness for O(1) memory.
- **Token bucket** allows controlled bursts up to a capacity while
  enforcing a steady average rate.
- **Leaky bucket** smooths bursts into a steady outflow rather than
  allowing them.

Having all five, with the same interface and the same test shape,
turns "which rate limiter should I use" from an abstract question into
something you can actually compare: run the same test against each,
read the same benchmark table, see the same Retry-After semantics
worked out five different ways.

## Consequences

- Five algorithms x two backends (memory, Redis) is ten
  implementations to maintain instead of one or two.
- The shared `RateLimiter` interface has to be small enough that all
  five algorithms can express their result through it — see
  [0002](0002-memory-and-redis-backends.md) and the `Result` struct
  (`Allowed`, `Remaining`, `RetryAfter`, `Limit`).
- Retry-After means something different per algorithm (see the
  README's Response headers section) — that had to be worked out five
  times, not once.
