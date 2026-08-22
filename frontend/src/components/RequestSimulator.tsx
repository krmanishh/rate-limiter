"use client";

import { useRef, useState } from "react";
import { ApiError, checkRateLimit } from "@/lib/api";
import type { SimulatedRequest } from "@/lib/types";
import Button from "./ui/Button";
import Card from "./ui/Card";
import EmptyState from "./ui/EmptyState";
import Spinner from "./ui/Spinner";
import RequestTimeline from "./RequestTimeline";
import ResultsTable from "./ResultsTable";

const DELAY_BETWEEN_REQUESTS_MS = 150;
const MAX_REQUESTS = 100;

export default function RequestSimulator() {
  const [apiKey, setApiKey] = useState("demo-user");
  const [count, setCount] = useState(10);
  const [running, setRunning] = useState(false);
  const [results, setResults] = useState<SimulatedRequest[]>([]);
  const stopRequested = useRef(false);

  const trimmedKey = apiKey.trim();
  const canRun = !running && trimmedKey.length > 0;

  async function run() {
    if (!trimmedKey) return;

    setRunning(true);
    stopRequested.current = false;
    setResults([]);

    const total = Math.min(Math.max(Math.floor(count) || 0, 1), MAX_REQUESTS);

    for (let i = 1; i <= total; i++) {
      if (stopRequested.current) break;

      const timestamp = new Date();

      try {
        const response = await checkRateLimit(trimmedKey);

        setResults((previous) => [
          ...previous,
          {
            id: i,
            timestamp,
            allowed: response.allowed,
            remaining: response.remaining,
            retryAfter: response.retry_after,
          },
        ]);
      } catch (err) {
        setResults((previous) => [
          ...previous,
          {
            id: i,
            timestamp,
            allowed: false,
            remaining: 0,
            retryAfter: 0,
            error: err instanceof ApiError ? err.message : "Request failed",
          },
        ]);

        // A request-level failure (network down, API unreachable) will
        // almost certainly repeat for every remaining request — stop
        // rather than hammering a dead API `total` times.
        break;
      }

      if (i < total) {
        await new Promise((resolve) => setTimeout(resolve, DELAY_BETWEEN_REQUESTS_MS));
      }
    }

    setRunning(false);
  }

  function stop() {
    stopRequested.current = true;
  }

  return (
    <Card
      title="Request simulator"
      description="Sends requests to POST /api/v1/ratelimit/check."
    >
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
        <div className="flex-1">
          <label
            htmlFor="api-key"
            className="mb-1 block text-xs font-medium text-slate-500"
          >
            API key / client identifier
          </label>
          <input
            id="api-key"
            value={apiKey}
            onChange={(event) => setApiKey(event.target.value)}
            disabled={running}
            placeholder="e.g. demo-user"
            className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm disabled:opacity-60 dark:border-slate-700 dark:bg-slate-800"
          />
        </div>

        <div className="w-full sm:w-36">
          <label
            htmlFor="request-count"
            className="mb-1 block text-xs font-medium text-slate-500"
          >
            # requests (max {MAX_REQUESTS})
          </label>
          <input
            id="request-count"
            type="number"
            min={1}
            max={MAX_REQUESTS}
            value={count}
            onChange={(event) => setCount(Number(event.target.value))}
            disabled={running}
            className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm disabled:opacity-60 dark:border-slate-700 dark:bg-slate-800"
          />
        </div>

        <div>
          {running ? (
            <Button variant="danger" onClick={stop}>
              Stop
            </Button>
          ) : (
            <Button onClick={run} disabled={!canRun}>
              Run
            </Button>
          )}
        </div>
      </div>

      <div className="mt-6">
        {results.length === 0 && running && (
          <span className="inline-flex items-center gap-2 text-sm text-slate-500">
            <Spinner /> Sending first request…
          </span>
        )}

        {results.length === 0 && !running && (
          <EmptyState
            title="No requests sent yet"
            description="Enter a key and click Run to start."
          />
        )}

        {results.length > 0 && (
          <div className="space-y-4">
            <RequestTimeline results={results} />
            <ResultsTable results={results} />
          </div>
        )}
      </div>
    </Card>
  );
}
