# 6. Make the Redis failure policy configurable, default to fail-closed

## Status

Accepted

## Context

Redis can become unreachable — network partition, restart, overload.
When the Redis-backed limiter's Lua script call fails, the service has
to return *some* answer to "is this request allowed?" without knowing
what the real rate limit state actually is. There is no universally
correct answer:

- Reject the request ("fail closed"): safer for whatever the rate
  limit protects, but an outage now blocks 100% of traffic instead of
  0%.
- Allow the request ("fail open"): keeps the service available, but an
  outage means the rate limit — and whatever it exists to protect — is
  effectively disabled for as long as Redis is down.

## Decision

Make this configurable via `RATE_LIMIT_FAIL_MODE=closed|open`, default
`closed`. The decision lives entirely in `internal/limiter/adapter`
(`RedisAdapter`), the one place that can tell a genuine Redis error
apart from a legitimate rejection — the algorithm implementations
themselves never see this distinction.

## Why

"Which failure mode is correct" depends on what's being protected, and
that's a property of the deployment, not of this codebase. A rate
limiter guarding a fragile downstream payment processor should almost
certainly fail closed. A rate limiter guarding a cache-warming endpoint
where availability matters more than perfect enforcement might
reasonably fail open. Hard-coding either choice would be wrong for the
other use case.

Fail-closed as the default follows the more conservative, "safer by
default" convention: an operator has to explicitly opt into "let
traffic through when we can't verify the limit," rather than
discovering after the fact that an outage silently disabled protection.

Whichever mode is active, the failure is never silent: it's counted in
`rate_limit_errors_total` (distinct from `rate_limit_rejected_total`)
and logged via `slog.Warn` with the algorithm and mode — see
`internal/limiter/adapter/redis.go`. A wave of 429s during a real
outage should be diagnosable from metrics/logs as "Redis is down," not
mistaken for a wave of legitimately rate-limited traffic.

## Consequences

- The adapter has two responsibilities now — adapting the
  context/error-returning Redis limiter interface to the plain
  `RateLimiter` interface, and enforcing this policy on error. See
  [0008](0008-redis-adapter.md) for why that's one cohesive
  responsibility rather than two that should be split apart.
- Fail-open's `Result` can't report a true `Remaining` count (the real
  state is unknown during an outage); it reports the full configured
  limit as a best-effort value rather than inventing a number with
  false precision.
- This is tested against a real Redis outage, not a mock — see
  `internal/limiter/adapter/redis_failure_test.go`
  (`REDIS_FAILURE_TEST=1`), which stops and restarts the actual
  `rate-limiter-redis` container.
