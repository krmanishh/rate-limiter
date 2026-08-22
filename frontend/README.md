# Rate Limiter Playground (frontend)

A Next.js + TypeScript + Tailwind CSS dashboard for the Go rate
limiter API in the parent directory. It's a plain client — no server
of its own beyond Next.js itself, no database, no proxy route. Every
data-fetching component calls the Go API directly from the browser.

## What it shows

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

### CORS

The browser calls the Go API's origin directly, so the API has to
allow this app's origin via CORS. The API's `CORS_ALLOWED_ORIGIN`
defaults to `http://localhost:3000` — matching `next dev`'s default
port — so local development works with no extra configuration. If you
run this app on a different port, or deploy it somewhere, update
`CORS_ALLOWED_ORIGIN` on the API to match.

## Project structure

```
src/
  app/
    layout.tsx        Root layout, metadata, fonts
    page.tsx           Assembles the dashboard's sections
    globals.css        Tailwind entry point + CSS variables

  components/
    StatusPanel.tsx        Server status (GET /health, polled)
    ConfigPanel.tsx         Active configuration (GET /api/v1/config)
    AlgorithmExplorer.tsx   Educational algorithm selector + pros/cons
    RequestSimulator.tsx    The key/count form and run loop
    RequestTimeline.tsx     Colored-block visual timeline
    ResultsTable.tsx        Per-request results table
    ui/                     Small reusable primitives: Card, Badge,
                            Button, Spinner, EmptyState, ErrorState

  lib/
    types.ts            TypeScript types mirroring the Go API's JSON
                        contracts exactly (field names match the wire
                        format, e.g. retry_after not retryAfter)
    api.ts               Typed fetch wrappers + error handling
    algorithms.ts         Static metadata: descriptions, pros/cons, and
                          which config fields apply to each algorithm
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
