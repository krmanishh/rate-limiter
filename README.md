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
                         ┌──────────────────┐
                         │  Rate Middleware  │
                         └────────┬─────────┘
                                  ↓
                         ┌──────────────────┐
                         │  RateLimiter      │
                         │   interface       │
                         └────────┬─────────┘
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
| `RATE_LIMIT_LIMIT` | `5` | Requests per window (fixed window, sliding log, sliding counter) |
| `RATE_LIMIT_WINDOW` | `1m` | Window size, as a Go duration string (`30s`, `1m`, `2h`) |
| `RATE_LIMIT_CAPACITY` | `10` | Bucket capacity (token bucket, leaky bucket) |
| `RATE_LIMIT_REFILL_RATE` | `2` | Tokens added per second (token bucket) |
| `RATE_LIMIT_LEAK_RATE` | `2` | Requests drained per second (leaky bucket) |
| `REDIS_ADDR` | `localhost:6379` | `host:port` of the Redis server (only used when storage is `redis`) |
| `SERVER_PORT` | `8080` | HTTP server listen port |

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

Some limiter tests are integration tests that need a live Redis on
`localhost:6379` (see `docker-compose.yml`); they fail fast with a
clear error if Redis isn't reachable.

## Running with Docker

```bash
docker compose up --build
```

This starts two containers:

```
docker-compose up
        │
        ├── rate-limiter-api  :8080   (this service, built from Dockerfile)
        │
        └── rate-limiter-redis :6379  (redis:7-alpine)
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

## API endpoints

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/health` | Liveness check. Always `200 {"status":"ok"}`. |
| `GET` | `/api/v1/protected-resource` | Demo resource behind the rate limit middleware. Keyed by the `X-API-Key` header (falls back to remote IP). |
| `POST` | `/api/v1/ratelimit/check` | Ad-hoc rate decision for an arbitrary key, independent of any specific resource — useful if another service wants to ask "would this be allowed?" without being the protected resource itself. |

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

## Response headers

Set by the rate limit middleware on every request it guards:

| Header | Meaning |
|---|---|
| `X-RateLimit-Limit` | The configured limit/capacity for this key. |
| `X-RateLimit-Remaining` | Requests remaining in the current window/bucket. |
| `Retry-After` | Seconds until the request would be allowed again. Set on rejection for fixed window, sliding log, and token bucket. Sliding counter and leaky bucket don't currently compute a retry time, so they omit it even when rejecting. |

## Project structure

```
cmd/server/            Entry point: config → factory → router → HTTP server, with
                        graceful shutdown

internal/config/       Environment-variable loading (RateLimitConfig, ServerConfig)

internal/model/        HTTP request/response DTOs

internal/api/          HTTP handlers, router, full HTTP integration tests

internal/middleware/   Rate limit middleware (extracts key, calls the limiter,
                        sets response headers)

internal/limiter/
  limiter.go            RateLimiter interface + Result
  fixedwindow/          Fixed window: memory (limiter.go) + Redis (redis.go)
  slidinglog/           Sliding log:  memory + Redis
  slidingcounter/       Sliding counter: memory + Redis
  tokenbucket/          Token bucket: memory + Redis
  leakybucket/          Leaky bucket: memory + Redis
  redis/                Shared helper for parsing Lua script results
  adapter/              Adapts the context-aware Redis limiter interface to
                        the plain RateLimiter interface
  factory/              Builds a RateLimiter (+ io.Closer) from RateLimitConfig

internal/store/
  store.go              Storage interface used by the Redis-backed limiters
  redisstore/            go-redis implementation of that interface

Dockerfile              Multi-stage build for the API container
docker-compose.yml       API + Redis, wired together for local/dev use
```
