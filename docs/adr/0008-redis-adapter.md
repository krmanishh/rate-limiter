# 8. Adapt the Redis limiter interface instead of widening RateLimiter

## Status

Accepted

## Context

Every Redis-backed algorithm needs a `context.Context` (for
cancellation/timeouts on the network call) and needs to return an
`error` (the network call can fail) — its natural interface is
`Allow(ctx, key) (Result, error)`. Every memory-backed algorithm needs
neither: `Allow(key) Result` is the whole story, since there's no I/O
that can fail or needs cancelling. The rest of the codebase
(`middleware`, `api.Handler`) wants to call `Allow` without caring
which kind of limiter it's holding.

Two ways to reconcile this: change the shared `RateLimiter` interface
to `Allow(ctx, key) (Result, error)` everywhere, so memory
implementations always return `nil` for an error they can never
actually produce — or keep `RateLimiter` as `Allow(key) Result` and
adapt the Redis shape down to it at the boundary.

## Decision

Keep `RateLimiter.Allow(key) Result` as the single interface every
caller depends on. Introduce `internal/limiter/adapter.RedisAdapter`,
which wraps a `redislimiter.RateLimiter` (the `ctx`/`error`-returning
shape) and implements `RateLimiter` by supplying a background context
and resolving any error into a `Result` according to the configured
fail-mode (see [0006](0006-fail-open-vs-fail-closed.md)).

## Why

Widening `RateLimiter` to carry a context and an error would mean
every caller — `middleware`, `api.Handler`, every test that calls
`.Allow(key)` directly — has to pass a context and handle an error
that, for five of the ten concrete implementations, can never actually
occur. That's a real interface widened for a distinction (can this call
fail?) that only matters to one backend. It would also push the
question "what does it mean when a rate limit check fails" out to
every call site, instead of answering it once.

The adapter keeps that question — and its answer, the fail-mode policy
— in exactly one place. Callers stay simple: they hold a `RateLimiter`,
they call `Allow(key)`, they get a `Result`. Whether that `Result` came
from an in-process map lookup or a Redis round trip that happened to
fail and got resolved by policy is not something they need to know.

## Consequences

- The adapter has two responsibilities that could be seen as separate
  (shape adaptation; fail-mode policy) but are treated as one cohesive
  responsibility here: "be the boundary where Redis's failure modes get
  turned into a `Result` the rest of the system understands." Splitting
  them further would mean passing the fail-mode decision back out
  through another layer for no real benefit.
- `rate_limit_errors_total` is recorded here, not in `middleware` —
  the metrics architecture (see the README's Observability section)
  otherwise centralizes rate-limit metrics in the middleware, and this
  is the one deliberate exception, because the middleware never sees
  the underlying error at all.
- Only `internal/limiter/adapter` needs to know both shapes exist;
  `redislimiter.RateLimiter` and the plain `RateLimiter` never need to
  be reconciled anywhere else.
