# rate-limiter

A rate limiting HTTP service written in Go. It implements five classic
rate limiting algorithms behind a single interface, backed by either
in-process memory or Redis (for sharing limits across multiple
instances), and exposes them over a small HTTP API with a
rate-limiting middleware.

## What it is

- A `RateLimiter` interface with five algorithm implementations.
- Each algorithm has two backends: **memory** (single process) and
  **Redis** (shared across processes, via atomic Lua scripts).
- An HTTP server that uses a configured limiter to protect a resource,
  plus an ad-hoc "check" endpoint for querying the limiter directly.
- Everything is selected and tuned via environment variables — no code
  changes needed to switch algorithm, backend, or limits.

## Architecture

```
                         ┌──────────────────┐
                         │   HTTP Request   │
                         └────────┬─────────┘
                                  ↓
                         ┌──────────────────┐
                         │      Router      │   /health
                         └────────┬─────────┘   /api/v1/ratelimit/check
                                  ↓              /api/v1/protected-resource
                         ┌──────────────────┐    /metrics
                         │  Rate Middleware  │
                         └────────┬─────────┘
                        ┌─────────┴─────────┐
                        ↓                   ↓
                  Prometheus           RateLimiter
                    Metrics             interface
                                            ↓
                    ┌─────────────┴─────────────┐
                    ↓                           ↓
              Memory Limiter              Redis Adapter
              (per-process map)                 ↓
                                          Redis Limiter
                                          (Lua script)
                                                ↓
                                          Redis Store
                                                ↓
                                              Redis
```

Every route is also wrapped with an HTTP metrics middleware (`method`/`route`/`status`
labels), applied uniformly by the router — separate from the rate-limit-specific
metrics the middleware above records. Neither one lives inside the algorithm
implementations: `FixedWindow`, `TokenBucket`, etc. never import Prometheus.

Configuration flows the other direction, at startup only:

```
Environment
     ↓
config.Load() / config.LoadServer()
     ↓
RateLimitConfig
     ↓
factory.Create()
     ↓
RateLimiter + io.Closer
```

`factory.Create` returns the limiter *and* an `io.Closer`. For memory
storage that's a no-op; for Redis storage it's the Redis client, so
shutdown can close it uniformly without caring which backend is
active.

Shutdown:

```
Ctrl+C / SIGTERM
       ↓
signal.NotifyContext wakes up
       ↓
server.Shutdown() — stop accepting new requests,
                    let in-flight requests finish
       ↓
closer.Close() — close the Redis client (no-op for memory)
       ↓
process exits
```

## Supported algorithms

| Algorithm | Package | Idea |
|---|---|---|
| Fixed Window | `internal/limiter/fixedwindow` | Count requests in a fixed-size time bucket; reset when the bucket rolls over. Simple, but allows bursts at window boundaries. |
| Sliding Log | `internal/limiter/slidinglog` | Keep a timestamp per request; count how many fall inside the trailing window. Exact, but memory scales with request volume. |
| Sliding Counter | `internal/limiter/slidingcounter` | Weighted average of the current and previous fixed windows, approximating a sliding window cheaply. |
| Token Bucket | `internal/limiter/tokenbucket` | A bucket refills at a fixed rate; each request consumes a token. Allows controlled bursts up to capacity. |
| Leaky Bucket | `internal/limiter/leakybucket` | A queue drains at a fixed rate; each request adds to the queue. Smooths bursts into a steady outflow. |

Every algorithm implements:

```go
type RateLimiter interface {
    Allow(key string) Result
}

type Result struct {
    Allowed    bool
    Remaining  int
    RetryAfter int // seconds, only meaningful when Allowed == false
    Limit      int // the configured limit/capacity
}
```

## Memory vs Redis

| | Memory | Redis |
|---|---|---|
| Scope | One process | Shared across all instances talking to the same Redis |
| Setup | None | Requires a reachable Redis server |
| Consistency under concurrency | `sync.Mutex` per limiter | Atomic Lua script per `Allow()` call — no read-modify-write race |
| Clock source | Local wall clock | Redis server time (`TIME` command), so multiple app instances agree even under clock skew |
| State survives restart | No | Yes (until the algorithm's own TTL expires it) |

Select the backend with `RATE_LIMIT_STORAGE=memory` (default) or
`RATE_LIMIT_STORAGE=redis`.

## Configuration

All configuration is environment variables, loaded once at startup
(`internal/config`). Anything unset falls back to its default; invalid
values (bad enum, non-numeric field) fail startup with a clear error
instead of silently misbehaving.

| Variable | Default | Meaning |
|---|---|---|
| `RATE_LIMIT_ALGORITHM` | `fixed_window` | `fixed_window` \| `sliding_log` \| `sliding_counter` \| `token_bucket` \| `leaky_bucket` |
| `RATE_LIMIT_STORAGE` | `memory` | `memory` \| `redis` |
| `RATE_LIMIT_FAIL_MODE` | `closed` | `closed` \| `open` — what happens when Redis errors (see [Failure behavior](#failure-behavior-fail-open-vs-fail-closed)) |
| `RATE_LIMIT_LIMIT` | `5` | Requests per window (fixed window, sliding log, sliding counter) |
| `RATE_LIMIT_WINDOW` | `1m` | Window size, as a Go duration string (`30s`, `1m`, `2h`) |
| `RATE_LIMIT_CAPACITY` | `10` | Bucket capacity (token bucket, leaky bucket) |
| `RATE_LIMIT_REFILL_RATE` | `2` | Tokens added per second (token bucket) |
| `RATE_LIMIT_LEAK_RATE` | `2` | Requests drained per second (leaky bucket) |
| `REDIS_ADDR` | `localhost:6379` | `host:port` of the Redis server (only used when storage is `redis`) |
| `REDIS_PASSWORD` | `` (none) | Redis AUTH password. Empty means no AUTH — fine locally, set this in any real deployment. |
| `SERVER_PORT` | `8080` | HTTP server listen port |
| `CORS_ALLOWED_ORIGIN` | `http://localhost:3000` | Origin allowed to call this API from a browser (the frontend's `next dev` origin by default) |

### Failure behavior: fail-open vs fail-closed

`RATE_LIMIT_FAIL_MODE` only matters for Redis storage (memory limiters
can't fail). If Redis becomes unreachable mid-traffic:

- **`closed`** (default): reject the request (`429`, with a fixed
  5-second `Retry-After` since the real timing can't be known without a
  working Redis). Safer when the rate limit is protecting something
  expensive or fragile — an outage degrades to "nothing gets through"
  rather than silently removing the limit.
- **`open`**: allow the request. Prioritizes availability — an outage
  degrades to "unlimited" rather than "fully blocked."

Either way, the error is counted in `rate_limit_errors_total` (distinct
from a legitimate rejection) and logged via `slog.Warn`, so an outage
shows up in metrics/logs rather than just looking like a wave of normal
rate limiting.

## Running locally

Requires Go 1.26+.

```bash
go build ./...
go test ./...
```

Memory storage, no external dependencies:

```bash
RATE_LIMIT_ALGORITHM=fixed_window \
RATE_LIMIT_STORAGE=memory \
RATE_LIMIT_LIMIT=5 \
go run ./cmd/server
```

Redis storage — start Redis first (`docker compose up -d redis`, or any
local Redis on `localhost:6379`):

```bash
RATE_LIMIT_ALGORITHM=fixed_window \
RATE_LIMIT_STORAGE=redis \
RATE_LIMIT_LIMIT=5 \
REDIS_ADDR=localhost:6379 \
go run ./cmd/server
```

See [Testing](#testing) for what `go test ./...` actually covers and
what it needs (a live Redis for some suites).

## Running with Docker

```bash
docker compose up --build
```

This starts three containers:

```
docker-compose up
        │
        ├── rate-limiter-api        :8080  (this service, built from Dockerfile)
        │
        ├── rate-limiter-redis      :6379  (redis:7-alpine)
        │
        └── rate-limiter-prometheus :9090  (scrapes the API's /metrics)
```

The API container is configured with `REDIS_ADDR=redis:6379` —
**`redis` is the Compose service name**, not `localhost`. Inside the
`rate-limiter` container, `localhost` refers to that container itself;
to reach the other container you address it by its service name, which
Docker Compose resolves on the shared network. This is why
`docker-compose.yml` sets `REDIS_ADDR=redis:6379` rather than
`REDIS_ADDR=localhost:6379`.

Verify it's up:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

Exercise the rate limit (default `RATE_LIMIT_LIMIT=5` in
`docker-compose.yml`):

```bash
for i in $(seq 1 7); do
  curl -s -o /dev/null -w "%{http_code}\n" \
    http://localhost:8080/api/v1/protected-resource \
    -H "X-API-Key: $CLIENT_ID"
done
# 200 200 200 200 200 429 429
```

The Go binary is built in a multi-stage `Dockerfile` (`golang:1.26-alpine`
builder → `alpine:3.20` runtime, `CGO_ENABLED=0`, stripped binary,
non-root user), producing a small (~25 MB) production image.

Graceful shutdown works the same way in a container: `docker stop`
sends `SIGTERM`, which the server catches to finish in-flight requests
and close its Redis connection before exiting.

## Frontend

[`frontend/`](frontend/README.md) is a Next.js + TypeScript + Tailwind
CSS "Rate Limiter Playground" — a dashboard that talks to this API
directly from the browser: server status, the active configuration,
an algorithm comparison, and a request simulator with a live timeline.
It's a separate app with its own `package.json`; it doesn't share a
process or a port with the Go server.

```bash
# terminal 1: the Go API (memory storage, no Redis needed to try the UI)
go run ./cmd/server

# terminal 2: the frontend
cd frontend && npm install && npm run dev
```

Then open `http://localhost:3000`. The API's default
`CORS_ALLOWED_ORIGIN` already matches `next dev`'s default port, so no
extra configuration is needed for local development. See
[`frontend/README.md`](frontend/README.md) for details.

## Testing

```bash
go test ./...              # everything below except the two opt-in suites
go test -race ./...        # same, with the race detector
go test -bench=. -benchmem -run=^$ ./internal/limiter/...   # benchmarks (see Performance)
```

What actually runs:

- **Unit tests** for every memory-backed algorithm, config loading, and
  HTTP handlers — no external dependencies.
- **Redis integration tests** for every Redis-backed algorithm and the
  factory — need a live Redis on `localhost:6379` (`docker compose up -d redis`).
  They fail fast with a clear error if Redis isn't reachable, rather
  than hanging or silently skipping.
- **Full HTTP integration tests** (`internal/api/integration_test.go`)
  that drive real requests through `httptest.NewServer(router)` —
  router → middleware → RateLimiter → response — for both memory and
  Redis storage.
- **Concurrency tests** for every Redis algorithm, firing more requests
  than the limit at once and asserting exactly `limit` succeed — proof
  the Lua scripts are atomic, not just "usually correct."
- **A distributed correctness test**
  (`TestCreate_DistributedAcrossInstances`) that builds three
  independent limiters against the same Redis, hammers the same key
  from all three concurrently, and asserts the limit is enforced once,
  not once per instance. This is the actual payoff of the Redis
  backend — see [ADR 0002](docs/adr/0002-memory-and-redis-backends.md).

Two suites are **not** part of `go test ./...`, on purpose:

- **Redis failure/recovery tests**
  (`internal/limiter/adapter/redis_failure_test.go`) stop and restart
  the real `rate-limiter-redis` container to verify fail-open/fail-closed
  behavior against an actual outage, then verify recovery. Opt in
  explicitly, since it disrupts shared infrastructure other tests use:

  ```bash
  REDIS_FAILURE_TEST=1 go test ./internal/limiter/adapter/... -run Failure -v
  ```

- **Benchmarks** run separately (`-bench=.`), since `go test` doesn't
  run them by default.

## CI

GitHub Actions (`.github/workflows/ci.yml`) runs on every push to
`main` and every pull request, as three independent jobs:

| Job | What it does |
|---|---|
| **Build & Test** | `gofmt` check, `go vet`, `go build`, `go test ./...`, `go test -race ./...` — against a real `redis:7-alpine` service container, so the Redis integration tests actually run in CI too. |
| **Lint** | `golangci-lint` (see below). |
| **Vulnerability check** | `govulncheck ./...` — catches known vulnerabilities in both dependencies and the Go standard library/toolchain itself. |

```bash
# lint locally the same way CI does
golangci-lint run ./...
```

`.golangci.yml` enables errcheck, govet, staticcheck, unused,
ineffassign, gocritic, revive, misspell, gosec, unconvert, and unparam.
Test-file cleanup calls (`defer conn.Close()`) are excluded from
errcheck/gosec — a cleanup failure isn't a test failure — and revive's
doc-comment rules are excluded, since this project follows a
comment-sparse house style (comments explain non-obvious *why*, not
*what*) rather than requiring godoc on every exported symbol.

The two opt-in test suites above are intentionally **not** run in CI:
the failure test would fight with the shared Redis service container
other jobs/tests use, and benchmarks aren't correctness checks.

## API endpoints

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/health` | Liveness check. Always `200 {"status":"ok"}`. |
| `GET` | `/api/v1/protected-resource` | Demo resource behind the rate limit middleware. Keyed by the `X-API-Key` header (falls back to remote IP). |
| `POST` | `/api/v1/ratelimit/check` | Ad-hoc rate decision for an arbitrary key, independent of any specific resource — useful if another service wants to ask "would this be allowed?" without being the protected resource itself. Always returns `200`; the decision is in the body (`allowed`/`remaining`/`retry_after`), not the status code. |
| `GET` | `/api/v1/config` | The server's active rate limit configuration (algorithm, storage, fail mode, and the relevant limit/window/capacity/rate fields) — lets a client render current config without separate access to the server's environment. Redis connection details are never included. |
| `GET` | `/metrics` | Prometheus exposition format. |

## Example requests

Protected resource, allowed:

```bash
curl -i http://localhost:8080/api/v1/protected-resource \
  -H "X-API-Key: $CLIENT_ID"
```

```
HTTP/1.1 200 OK
X-Ratelimit-Limit: 5
X-Ratelimit-Remaining: 4
Content-Type: application/json

{"message":"protected resource accessed"}
```

Protected resource, rejected after the limit is exceeded:

```
HTTP/1.1 429 Too Many Requests
X-Ratelimit-Limit: 5
X-Ratelimit-Remaining: 0
Retry-After: 47
Content-Type: application/json

{"error":"rate limit exceeded"}
```

Ad-hoc check:

```bash
curl -s -X POST http://localhost:8080/api/v1/ratelimit/check \
  -H "Content-Type: application/json" \
  -d '{"key":"some-key"}'
```

```json
{"allowed":true,"remaining":4,"retry_after":0}
```

Active configuration:

```bash
curl -s http://localhost:8080/api/v1/config
```

```json
{"algorithm":"fixed_window","storage":"memory","fail_mode":"closed","limit":5,"window_seconds":60,"capacity":10,"refill_rate":2,"leak_rate":2}
```

## Response headers

Set by the rate limit middleware on every request it guards:

| Header | Meaning |
|---|---|
| `X-RateLimit-Limit` | The configured limit/capacity for this key. |
| `X-RateLimit-Remaining` | Requests remaining in the current window/bucket. |
| `Retry-After` | Seconds until the request would be allowed again, computed per-algorithm (see below). Always set on rejection. |

`Retry-After` means something specific to each algorithm's own mechanics, not a generic "try again later":

| Algorithm | Retry-After is... |
|---|---|
| Fixed Window | Time until the current window expires. |
| Sliding Log | Time until the oldest logged request ages out of the window. |
| Sliding Counter | *Estimated* time until the weighted current+previous count decays below the limit — an estimate because the algorithm itself only approximates a true sliding window. |
| Token Bucket | Time until enough tokens refill for one more request. |
| Leaky Bucket | Time until the queue leaks enough to have a free slot. |

Sub-second waits are rounded **up** to whole seconds (never down) — reporting `Retry-After: 0` would tell a client to retry immediately when it actually can't yet.

## Observability

**Logs** are structured JSON (`log/slog`, stdlib only — no logging
dependency), written to stdout. A rejected request logs something like:

```json
{"time":"...","level":"INFO","msg":"request rate limited","key":"api-key:user-123","algorithm":"fixed_window","storage":"redis","remaining":0}
```

(Only the server binary sets the JSON handler; running tests directly
uses Go's default text handler, which is why test output looks
different.)

**Metrics** are exposed at `/metrics` in Prometheus format:

| Metric | Type | Labels | Meaning |
|---|---|---|---|
| `rate_limit_requests_total` | counter | `algorithm`, `storage` | Every rate limit decision made. |
| `rate_limit_allowed_total` | counter | `algorithm`, `storage` | Decisions that allowed the request. |
| `rate_limit_rejected_total` | counter | `algorithm`, `storage` | Decisions that rejected the request. |
| `rate_limit_errors_total` | counter | `algorithm`, `storage` | Failures to reach a decision at all (e.g. Redis unreachable) — distinct from a legitimate rejection. Only the Redis path can produce these; memory limiters never error. |
| `http_requests_total` | counter | `method`, `route`, `status` | Every HTTP request served. |
| `http_request_duration_seconds` | histogram | `method`, `route`, `status` | Request latency. |

```bash
curl http://localhost:8080/metrics
```

Bring up Prometheus alongside the stack (scrapes the API every 5s,
config in `prometheus.yml`):

```bash
docker compose up --build
```

```
docker compose
│
├── rate-limiter   :8080
├── redis          :6379
└── prometheus     :9090
```

Prometheus's UI is at `http://localhost:9090` — check
**Status → Targets** to confirm the `rate-limiter` job is `UP`, or query
`rate_limit_requests_total` directly. No Grafana yet; Prometheus's own
UI is enough to start with.

## Performance

```bash
go test -bench=. -benchmem -run=^$ ./internal/limiter/...
```

Measured on an Apple M3 Pro, single key, steady state (limit high enough
that nothing rejects, so this measures the "allowed" hot path only):

| Algorithm | Memory | Redis | Redis is roughly... |
|---|---|---|---|
| Fixed Window | 40.6 ns/op, 0 allocs | 128,675 ns/op, 17 allocs | 3,170x slower |
| Sliding Log | 72.8 ns/op, 0 allocs | 139,030 ns/op, 21 allocs | 1,910x slower |
| Sliding Counter | 46.6 ns/op, 0 allocs | 141,102 ns/op, 18 allocs | 3,030x slower |
| Token Bucket | 52.4 ns/op, 0 allocs | 138,287 ns/op, 18 allocs | 2,640x slower |
| Leaky Bucket | 48.5 ns/op, 0 allocs | 140,290 ns/op, 18 allocs | 2,890x slower |

That's the actual answer to "what does Redis cost compared to local
memory": a rate limit decision goes from tens of nanoseconds (a mutex
lock and a map lookup) to ~130-140 microseconds (a real network round
trip plus Lua script execution) — roughly **2,000-3,000x**. That cost
buys correctness across multiple instances: with memory storage, each
app instance enforces its own separate limit (3 instances × a limit of
5 lets 15 requests through); with Redis, the limit is enforced on the
shared key regardless of how many instances are running. Whether that
trade is worth it depends entirely on whether you're running one
instance or several.

Numbers vary by machine and by how close the Redis server is (this was
measured against a local Docker container on loopback — a real network
hop to a managed Redis would add more). Re-run the benchmark on your
own hardware before trusting these figures for capacity planning.

## Security

**Client identification.** Requests are keyed by the `X-API-Key`
header, falling back to the connecting IP (with the ephemeral port
stripped — an earlier version of this code kept the port, which meant
the IP fallback never actually rate-limited anything, since no two
requests share a port). This deliberately does **not** consult
`X-Forwarded-For` or similar headers. Trusting a forwarded-for header
requires knowing this service sits behind a specific proxy that sets it
correctly and strips any client-supplied value first — this service has
no way to verify that on its own. Without that guarantee, any client
could set the header themselves and rate-limit as someone else, or
evade limits entirely by spoofing a fresh value per request. If this is
ever deployed behind a trusted reverse proxy, that proxy's real-IP
extraction should *replace* `clientIP()` in
`internal/middleware/ratelimit.go`, not be layered on top of it.

**Logging.** Rate-limit keys are hashed (truncated SHA-256) before
appearing in logs (`key_hash`, not `key`) — the raw value may embed a
caller-supplied API key, and log aggregation systems are generally a
less-trusted, longer-retained destination than you want a credential
ending up in. The hash still lets you correlate repeated hits from the
same key across log lines.

**Redis.** `REDIS_PASSWORD` enables AUTH; unset means no
authentication, which is fine for local development but not for a real
deployment. There's currently no TLS support for the Redis connection —
if Redis isn't reachable over a trusted private network, that's a real
gap. Redis's own port (6379) should never be exposed to the public
internet regardless of AUTH; you don't want it in your Compose port
mapping in production, only in an internal network.

**Request size limits.** `/api/v1/ratelimit/check`'s body is capped at
1 KiB (`http.MaxBytesReader`) — the only expected payload is
`{"key":"..."}`, so there's no reason to buffer an arbitrarily large
body before JSON decoding fails on it anyway.

**Timeouts.** The HTTP server sets `ReadHeaderTimeout`, `ReadTimeout`,
`WriteTimeout`, and `IdleTimeout` — standard hardening against
slow-client resource exhaustion (slowloris-style attacks), where a
client opens a connection and trickles bytes in just fast enough to
keep it alive indefinitely.

**CORS.** Not implemented, deliberately — this is a server-to-server
API (rate limiting decisions and a protected resource demo), not
something a browser is expected to call directly with `fetch()`/XHR
from a third-party origin. Add CORS middleware only if a browser client
actually needs one.

**Container.** Runs as a non-root user in a minimal Alpine image (see
[Running with Docker](#running-with-docker)) — a compromised process
inside the container doesn't get root by default.

**Dependencies.** CI runs [`govulncheck`](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
on every push, which checks both third-party dependencies and the Go
standard library/toolchain itself against the official Go vulnerability
database — a stdlib CVE (e.g. in `crypto/tls` or `net/http`) is exactly
as real a risk to this service as one in `go-redis`. Run it locally:

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## Architecture Decision Records

The design choices behind this project — and, as important, the
alternatives that were considered and why they weren't picked — are
recorded in [`docs/adr`](docs/adr/README.md):

- [0001](docs/adr/0001-five-rate-limiting-algorithms.md) — Why five algorithms, not one
- [0002](docs/adr/0002-memory-and-redis-backends.md) — Why both memory and Redis
- [0003](docs/adr/0003-why-redis.md) — Why Redis specifically
- [0004](docs/adr/0004-lua-scripts-for-atomicity.md) — Why Lua scripts
- [0005](docs/adr/0005-server-side-redis-time.md) — Why server-side Redis TIME
- [0006](docs/adr/0006-fail-open-vs-fail-closed.md) — Why a configurable fail-open/fail-closed policy
- [0007](docs/adr/0007-factory-pattern.md) — Why a factory
- [0008](docs/adr/0008-redis-adapter.md) — Why an adapter instead of widening the interface

## Project structure

```
cmd/server/            Entry point: config → factory → router → HTTP server, with
                        graceful shutdown

internal/config/       Environment-variable loading (RateLimitConfig, ServerConfig)

internal/model/        HTTP request/response DTOs

internal/api/          HTTP handlers, router, full HTTP integration tests

internal/middleware/   Rate limit middleware (extracts key, calls the limiter,
                        sets response headers, records metrics/logs) and a
                        generic HTTP metrics middleware used by the router

internal/metrics/      Prometheus metric definitions — the only package that
                        imports the Prometheus client

internal/limiter/
  limiter.go            RateLimiter interface + Result
  fixedwindow/          Fixed window: memory (limiter.go) + Redis (redis.go)
  slidinglog/           Sliding log:  memory + Redis
  slidingcounter/       Sliding counter: memory + Redis
  tokenbucket/          Token bucket: memory + Redis
  leakybucket/          Leaky bucket: memory + Redis
  redis/                The Redis-limiter interface contract (ctx + error) and
                        shared helpers for parsing Lua script results
  adapter/              Adapts that interface to the plain RateLimiter
                        interface; owns the fail-open/fail-closed policy and
                        rate_limit_errors_total on a Redis error
  factory/              Builds a RateLimiter (+ io.Closer) from RateLimitConfig

internal/store/
  store.go              Storage interface used by the Redis-backed limiters
  redisstore/            go-redis implementation of that interface

Dockerfile              Multi-stage build for the API container
docker-compose.yml       API + Redis + Prometheus, wired together for local/dev use
prometheus.yml           Prometheus scrape config (targets the API's /metrics)
.github/workflows/       CI: build/test/race, lint, vulnerability check
docs/adr/                Architecture Decision Records
frontend/                Next.js + TypeScript + Tailwind dashboard (separate app,
                          see frontend/README.md)
```
