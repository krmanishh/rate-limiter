# rate-limiter

A rate limiting HTTP service in Go, with an interactive dashboard for
exploring it. The repository is split into two independent projects:

```
rate-limiter/
├── backend/    Go API — five rate limiting algorithms, memory or Redis
│               backed, metrics, structured logging, ADRs. Own go.mod.
│               See backend/README.md.
│
└── frontend/   Next.js + TypeScript + Tailwind dashboard that talks to
                the API directly from the browser. Own package.json.
                See frontend/README.md.
```

Each has its own README with full details (architecture, configuration,
testing, CI, security notes, ADRs for the backend). This file is just
the map.

## Running both together

```bash
# terminal 1 — the Go API (memory storage, no Redis needed to try the UI)
cd backend
go run ./cmd/server

# terminal 2 — the frontend
cd frontend
npm install
npm run dev
```

Open `http://localhost:3000`. The API's default `CORS_ALLOWED_ORIGIN`
already matches `next dev`'s port, so no extra configuration is needed
locally.

To run the API with Redis and Prometheus via Docker Compose instead:

```bash
cd backend
docker compose up --build
```

## Where to go next

- **Backend**: [`backend/README.md`](backend/README.md) — architecture,
  all five algorithms, configuration, Docker, testing, CI, security,
  and [Architecture Decision Records](backend/docs/adr/README.md).
- **Frontend**: [`frontend/README.md`](frontend/README.md) — what the
  dashboard shows and how it's built.

## Repository layout

```
backend/     Go module (cmd/, internal/, Dockerfile, docker-compose.yml,
             prometheus.yml, docs/adr/) — self-contained, buildable and
             testable on its own from within this directory.
frontend/    Next.js app (its own package.json) — likewise self-contained.
.github/     CI, spanning both (currently backend-only checks).
```
