# 4. Use Lua scripts for every Redis-backed algorithm

## Status

Accepted

## Context

Every algorithm's Redis implementation needs to read some state (a
count, a timestamp, a token balance), decide whether to allow the
request, and write updated state back — atomically. If those steps
happen as separate Redis commands from the application (e.g. `GET`
then `SET`), two concurrent requests can interleave between them: both
read count=4 against a limit of 5, both decide "allowed," both write
count=5 — and the limit was actually violated, because the true count
should have been 6 with one of them rejected.

## Decision

Implement every algorithm's Redis logic as a single Lua script
executed via `EVAL`, not as a sequence of separate commands from the
Go client.

## Why

Redis executes a Lua script as one atomic unit — no other command,
from any client, can run in between two lines of the script. That
turns the check-then-update race described above into a non-issue: the
whole "read state, compute the decision, write state" sequence happens
as one indivisible operation from Redis's point of view.

This is proven, not just asserted — every algorithm's Redis test suite
includes a concurrency test that fires more requests than the limit at
once and asserts exactly `limit` succeed (see e.g.
`TestRedisLimiter_Concurrent` in each algorithm's `redis_test.go`, and
the cross-instance version in
`TestCreate_DistributedAcrossInstances`).

The alternative — `INCR` then a separate `EXPIRE`, or a Go-side
read-modify-write using `GET`/`SET` — would need Redis transactions
(`MULTI`/`EXEC` with `WATCH`) to be safe, which is more complex to
reason about and doesn't extend cleanly to the more involved
algorithms (sliding counter's two-window weighting, token/leaky
bucket's time-based decay) the way a script does.

## Consequences

- Each algorithm's logic exists in two places conceptually — the Go
  memory implementation and the Lua string in `redis.go` — that have to
  stay behaviorally consistent. They're tested separately but should
  agree on outcomes for the same inputs.
- Lua scripts are harder to unit test in isolation than Go code; they're
  exercised indirectly through the Go-level Redis integration tests
  instead.
- Debugging a Lua script's logic means reading Lua, not Go — a genuine
  context switch, though the scripts are kept short and commented.
