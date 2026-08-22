# Rate Limiter Playground (frontend)

A Next.js + TypeScript + Tailwind CSS dashboard for the Go rate
limiter API in the parent directory. It's a plain client — no server
of its own beyond Next.js itself, no database, no proxy route. Every
data-fetching component calls the Go API (or Prometheus) directly
from the browser.

It has two pages:

- **`/` — Rate Limiter Playground** (Phase 1): server status, live
  configuration, an algorithm explorer, and a request simulator.
- **`/observability` — Observability Dashboard** (Phase 2): charts
  and KPIs built from the metrics the Go API already exposes at
  `/metrics`, queried through Prometheus.

## Playground (`/`)

### What it shows

- **Server status** — polls `GET /health`.
- **Current configuration** — `GET /api/v1/config`: the algorithm,
  storage backend, fail mode, and whichever limit/window/capacity/rate
  fields are relevant to that algorithm.
- **Algorithm explorer** — a selector across all five algorithms
  (fixed window, sliding log, sliding counter, token bucket, leaky
  bucket) with a description and pros/cons for each. This is purely
  informational: it does not change what the live backend is running.
  The backend's algorithm is chosen at its own startup via
  `RATE_LIMIT_ALGORITHM` and can't be changed at runtime — see
  [`../backend/README.md`](../backend/README.md)'s Configuration section.
- **Request simulator** — enter a key and a request count, and it
  fires that many requests at `POST /api/v1/ratelimit/check`
  sequentially (with a small delay between each so the timeline is
  watchable), rendering a colored timeline and a results table
  (request #, timestamp, allowed/rejected, remaining, retry-after) as
  responses come back.

## Observability Dashboard (`/observability`)

Everything on this page is read directly from Prometheus's HTTP query
API (`/api/v1/query`, `/api/v1/query_range`), not from the Go API and
not recomputed in the frontend. Prometheus is already scraping the Go
API's `/metrics` endpoint (see
[`../backend/prometheus.yml`](../backend/prometheus.yml)) and storing
the history this page charts — the frontend just queries it. Its
default CORS policy (`Access-Control-Allow-Origin: *`) is what makes
calling it directly from the browser possible with no backend changes.

Every metric shown maps to one already defined in
[`../backend/internal/metrics/metrics.go`](../backend/internal/metrics/metrics.go):

- **Total / allowed / rejected requests, rejection rate** —
  `rate_limit_requests_total`, `rate_limit_allowed_total`,
  `rate_limit_rejected_total`.
- **Requests over time, allowed vs. rejected** — `rate()` over the
  same three counters, charted across the selected time range.
- **Algorithm usage** — `rate_limit_requests_total` grouped by its
  `algorithm` and `storage` labels.
- **HTTP latency (p50 / p95)** — `histogram_quantile()` over
  `http_request_duration_seconds`. There is no metric that times the
  rate-limit check in isolation, so this is labeled as overall HTTP
  latency, not a limiter-only number.
- **Redis health** — there's no dedicated Redis probe (no
  `redis_exporter` in the stack), so this is inferred from
  `rate_limit_errors_total`'s recent rate plus the active storage
  backend from `GET /api/v1/config`, and clearly labeled "inferred"
  in the UI. It reads "N/A" when the active backend isn't Redis.
- **Backend health** — reuses the Playground's existing
  `StatusPanel` component (`GET /health`) rather than duplicating it.

Time range (5m/15m/1h/6h) and auto-refresh interval (off/10s/30s/60s)
are both adjustable from the page; a manual "Refresh now" button and
a last-updated timestamp are always available. If Prometheus is
unreachable, the page shows a single dashboard-wide error with a
retry button rather than N separately-broken tiles.

## Running it

Requires Node 20+.

```bash
npm install
cp .env.example .env.local   # only if you need a different API URL
npm run dev
```

Then open `http://localhost:3000`. The Go API needs to be running
separately — see [`../backend/README.md`](../backend/README.md), or
the [repository root README](../README.md) for the two-terminal
quickstart — this app doesn't start it.

### Configuration

| Variable | Default | Meaning |
|---|---|---|
| `NEXT_PUBLIC_API_BASE_URL` | `http://localhost:8080` | Base URL of the Go API. |
| `NEXT_PUBLIC_PROMETHEUS_BASE_URL` | `http://localhost:9090` | Base URL of Prometheus, used by the Observability Dashboard. |

### CORS

The browser calls the Go API's and Prometheus's origins directly, so
both have to allow this app's origin via CORS. The API's
`CORS_ALLOWED_ORIGIN` defaults to `http://localhost:3000` — matching
`next dev`'s default port — so local development works with no extra
configuration. If you run this app on a different port, or deploy it
somewhere, update `CORS_ALLOWED_ORIGIN` on the API to match.
Prometheus allows all origins by default and needs no configuration
change.

## Project structure

```
src/
  app/
    layout.tsx            Root layout, metadata, fonts, top nav
    page.tsx               Playground page (Phase 1)
    globals.css            Tailwind entry point + CSS variables
    observability/
      page.tsx              Observability Dashboard page (Phase 2)

  components/
    StatusPanel.tsx        Server status (GET /health, polled) — reused
                            on both pages
    ConfigPanel.tsx         Active configuration (GET /api/v1/config)
    AlgorithmExplorer.tsx   Educational algorithm selector + pros/cons
    RequestSimulator.tsx    The key/count form and run loop
    RequestTimeline.tsx     Colored-block visual timeline
    ResultsTable.tsx        Per-request results table
    ui/                     Small reusable primitives: Card, Badge,
                            Button, Select, Spinner, EmptyState, ErrorState
    observability/
      MetricTile.tsx           A single KPI number inside a Card
      TimeRangeControl.tsx     Time range + refresh interval + manual refresh
      RequestsOverTimeChart.tsx
      AllowedVsRejectedChart.tsx
      AlgorithmUsageChart.tsx
      LatencyChart.tsx
      RedisHealthTile.tsx      Inferred Redis health (see above)

  hooks/
    useObservabilityData.ts  Fetches everything the dashboard needs from
                              Prometheus in one batch; the single source
                              of truth so chart/tile components stay pure

  lib/
    types.ts            TypeScript types mirroring the Go API's JSON
                        contracts exactly (field names match the wire
                        format, e.g. retry_after not retryAfter)
    api.ts               Typed fetch wrappers + error handling for the Go API
    algorithms.ts         Static metadata: descriptions, pros/cons, and
                          which config fields apply to each algorithm
    prometheus.ts          Thin client for Prometheus's HTTP query API
    observabilityQueries.ts  All PromQL used by the dashboard, plus time
                              range / refresh interval presets
    observabilityTransform.ts  Pure helpers converting Prometheus's
                                matrix results into chart-ready arrays
```

Every data-fetching component handles three states explicitly:
loading (a spinner), error (a message with a retry button), and empty
(e.g. "no requests sent yet") — there's no unstyled blank gap or
silent failure anywhere in the data flow.

## Verifying

```bash
npx tsc --noEmit   # typecheck
npm run lint        # eslint (flat config, eslint-config-next)
npm run build        # production build
```
