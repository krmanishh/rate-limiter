# Rate Limiter Admin Console (frontend)

A Next.js + TypeScript + Tailwind CSS admin console for the Go rate
limiter API in the parent directory. It's a plain client — no server
of its own beyond Next.js itself, no database, no proxy route. Every
data-fetching component calls the Go API (or Prometheus) directly
from the browser.

It has six pages:

- **`/` — Rate Limiter Playground** (Phase 1): server status, live
  configuration, an algorithm explorer, and a request simulator.
- **`/observability` — Observability Dashboard** (Phase 2): charts
  and KPIs built from the metrics the Go API already exposes at
  `/metrics`, queried through Prometheus.
- **`/config` — Configuration** (Phase 3): a read-only view of the
  server's active settings.
- **`/keys` — API Keys** (Phase 3): a session-local key tester and
  per-key request history.
- **`/algorithms` — Algorithms** (Phase 3): a static comparison table
  plus an interactive, purely client-side simulator of all five
  algorithms.
- **`/system` — System** (Phase 3): an architecture diagram and
  consolidated backend/Redis/Prometheus health.

Dark mode (a manual toggle in the nav bar, persisted to
`localStorage`, defaulting to the OS preference) applies across every
page.

### A note on Phase 3 and what the backend actually supports

The Go API has **no endpoints to change configuration at runtime** and
**no API key registry** — see `backend/internal/config/loader.go` (env
vars, read once at startup) and `backend/internal/middleware/ratelimit.go`
(`X-API-Key` is just an arbitrary string used to partition the rate
limit, never stored or listed). Rather than build UI that pretends to
write settings or list keys the backend doesn't track, Phase 3's pages
are honest about this:

- `/config` shows live values in **disabled** selects, captioned with
  the environment variable that actually controls each one, plus a
  banner stating there's no runtime write endpoint.
- `/keys` tests keys against the real `POST /api/v1/ratelimit/check`
  and tracks results **only in the current browser session** — it says
  so explicitly, since there's no backend-side registry to read from.
- `/algorithms`' simulator runs simplified ports of each algorithm's
  actual Go logic **in the browser**, labeled as a local simulation —
  the server only ever runs one algorithm at a time, so this is the
  only way to compare all five against an identical request pattern.

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

## Configuration (`/config`)

Fetches `GET /api/v1/config` once (with a manual Refresh) and renders
three cards — algorithm, storage backend, and Redis failure policy —
each with a disabled `<select>` pre-filled with the live value and a
caption naming the controlling environment variable. The failure
policy card also shows whether it's currently "in effect" (only
meaningful when storage is Redis — see `FailMode`'s doc comment in
`backend/internal/config/config.go`).

## API Keys (`/keys`)

A form sends real requests to `POST /api/v1/ratelimit/check` for a
given key; results are tracked per-key in memory (via
`hooks/useKeyInspector.ts`) for as long as the page stays open. The
overview table aggregates each tracked key's sent/allowed/rejected
counts; selecting a key reuses the Playground's `RequestTimeline` and
`ResultsTable` components to show its full history. Nothing here is
read from or written to a backend-side key registry, because one
doesn't exist.

## Algorithms (`/algorithms`)

- **Comparison table** — the same descriptions/pros/cons/parameters
  from `lib/algorithms.ts` (used by the Playground's Algorithm
  Explorer), laid out side by side for all five algorithms at once.
- **Simulator** — `lib/algorithmSimulation.ts` ports each algorithm's
  `Allow()` logic from `backend/internal/limiter/*/limiter.go` into
  TypeScript, then runs all five against the same synthetic sequence
  of request offsets (configurable count/interval, plus each
  algorithm's real parameters). Results render as five
  `RequestTimeline`s so the algorithms' different behaviors under
  identical load are visually comparable. This never touches the
  network — it's a pure, deterministic, in-browser calculation.

## System (`/system`)

- **Health row** — backend (`StatusPanel`, reused from the
  Playground), Prometheus (reachability inferred from a live query,
  since Prometheus's `/-/healthy` doesn't send CORS headers the way
  `/api/v1/*` does), and Redis (`RedisHealthTile`, reused from the
  Observability Dashboard, fed by a single lightweight query in
  `hooks/usePrometheusSignal.ts` rather than a second copy of that
  logic).
- **Architecture diagram** — a static, plain-HTML/CSS picture of the
  request path (client → API → rate limit middleware → store) and the
  observability path (API → `/metrics` → Prometheus → this dashboard),
  annotated with the live algorithm/storage from `/api/v1/config`.

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
    layout.tsx            Root layout, metadata, fonts, top nav, no-flash
                          dark mode init script
    page.tsx               Playground page (Phase 1)
    globals.css            Tailwind entry point + CSS variables + dark variant
    observability/page.tsx  Observability Dashboard page (Phase 2)
    config/page.tsx          Configuration page (Phase 3)
    keys/page.tsx             API Keys page (Phase 3)
    algorithms/page.tsx        Algorithms page (Phase 3)
    system/page.tsx             System page (Phase 3)

  components/
    NavBar.tsx              Top nav: active links, mobile menu, theme toggle
    ThemeToggle.tsx          Dark/light switch (localStorage + useSyncExternalStore)
    StatusPanel.tsx        Server status (GET /health, polled) — reused widely
    ConfigPanel.tsx         Active configuration (GET /api/v1/config)
    AlgorithmExplorer.tsx   Educational algorithm selector + pros/cons
    RequestSimulator.tsx    The key/count form and run loop
    RequestTimeline.tsx     Colored-block visual timeline (formatTime is
                            pluggable — the Algorithms simulator uses it
                            for ms-offsets instead of wall-clock time)
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
      RedisHealthTile.tsx      Inferred Redis health — reused on /system too
    config/
      ReadOnlySetting.tsx      Disabled <select> + "set via env var X" caption
      AlgorithmConfigCard.tsx
      StorageConfigCard.tsx
      FailPolicyConfigCard.tsx
    keys/
      KeyForm.tsx              Test-a-key form
      KeyTable.tsx             Tracked keys + per-row run/stop/remove
      KeyAnalyticsPanel.tsx    Reuses RequestTimeline/ResultsTable per key
    algorithms/
      ComparisonTable.tsx      Static side-by-side algorithm comparison
      SimulatorControls.tsx    Shared limiter params + request pattern inputs
      SimulatorResults.tsx     Runs all 5 simulations, renders 5 timelines
    system/
      ArchitectureDiagram.tsx  Static request/observability path diagram
      PrometheusHealthTile.tsx

  hooks/
    useObservabilityData.ts  Fetches everything the dashboard needs from
                              Prometheus in one batch; the single source
                              of truth so chart/tile components stay pure
    useConfig.ts               Shared GET /api/v1/config fetch (config + system pages)
    useKeyInspector.ts          Per-key session state + the real check() run loop
    usePrometheusSignal.ts      One lightweight query used for both Prometheus
                                reachability and the Redis-health inference

  lib/
    types.ts            TypeScript types mirroring the Go API's JSON
                        contracts exactly (field names match the wire
                        format, e.g. retry_after not retryAfter)
    api.ts               Typed fetch wrappers + error handling for the Go API
    algorithms.ts         Static metadata: descriptions, pros/cons, and
                          which config fields apply to each algorithm
    algorithmSimulation.ts  Pure, dependency-free ports of each algorithm's
                            real Go Allow() logic, for the /algorithms simulator
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
