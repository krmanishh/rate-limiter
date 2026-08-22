# 3. Use Redis as the distributed backend

## Status

Accepted

## Context

Once the decision is made to support a shared backend at all (see
[0002](0002-memory-and-redis-backends.md)), something has to actually
hold that shared state. Candidates include Redis, a relational
database, or a purpose-built rate-limiting service.

## Decision

Use Redis.

## Why

- **Speed.** A rate limit decision has to happen on every request, on
  the hot path. Redis is an in-memory store built for exactly this
  latency profile — a relational database would add meaningfully more
  overhead per call for no benefit here (there's no need for joins,
  durability guarantees, or complex queries; every operation is a
  single key read-modify-write).
- **Atomicity primitives that fit the problem.** Redis's `INCR`,
  sorted sets (`ZADD`/`ZREMRANGEBYSCORE`), hashes, and — critically —
  Lua scripting map directly onto what these algorithms need: atomic
  counters, atomic pruning of old entries, atomic multi-field
  read-modify-write. See [0004](0004-lua-scripts-for-atomicity.md).
- **Ubiquity.** Redis is already the default choice for this kind of
  shared, ephemeral, latency-sensitive state across the industry — the
  operational knowledge (how to run it, monitor it, secure it) is
  widely available, unlike a bespoke rate-limiting service.
- **TTLs for free.** Every algorithm's Redis state needs to expire
  eventually so idle keys don't accumulate forever. Redis's native
  `EXPIRE`/`PEXPIRE` handles this without any separate cleanup process.

## Consequences

- The project depends on an external service being available and
  reachable — see [0006](0006-fail-open-vs-fail-closed.md) for what
  happens when it isn't.
- Redis's own security posture becomes this project's problem too (see
  the README's Security section: AUTH via `REDIS_PASSWORD`, no public
  network exposure, no TLS support yet).
