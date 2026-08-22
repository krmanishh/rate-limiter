# Architecture Decision Records

Lightweight records of the decisions that shape this project — not
exhaustive, just the ones where a different choice was genuinely
plausible and worth explaining.

| ADR | Decision |
|---|---|
| [0001](0001-five-rate-limiting-algorithms.md) | Implement five rate limiting algorithms, not one |
| [0002](0002-memory-and-redis-backends.md) | Support both memory and Redis backends |
| [0003](0003-why-redis.md) | Use Redis as the distributed backend |
| [0004](0004-lua-scripts-for-atomicity.md) | Use Lua scripts for every Redis-backed algorithm |
| [0005](0005-server-side-redis-time.md) | Use Redis's server-side TIME instead of the client's wall clock |
| [0006](0006-fail-open-vs-fail-closed.md) | Make the Redis failure policy configurable, default to fail-closed |
| [0007](0007-factory-pattern.md) | Use a factory to build a RateLimiter from configuration |
| [0008](0008-redis-adapter.md) | Adapt the Redis limiter interface instead of widening RateLimiter |
