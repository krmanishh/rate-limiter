"use client";

import { useRef, useState } from "react";
import { ApiError, checkRateLimit } from "@/lib/api";
import type { SimulatedRequest } from "@/lib/types";

export interface TrackedKey {
  key: string;
  history: SimulatedRequest[];
  running: boolean;
}

const MAX_REQUESTS_PER_RUN = 100;
const DELAY_BETWEEN_REQUESTS_MS = 150;

/**
 * Tracks real POST /api/v1/ratelimit/check results per key, entirely in
 * memory for the current page session.
 *
 * The backend has no concept of a registered API key — the X-API-Key
 * header (see backend/internal/middleware/ratelimit.go's getRateLimitKey)
 * is just an arbitrary string used to partition the rate limit, and
 * there's no endpoint to list or query keys it has seen. So this isn't
 * a view onto some backend registry — it's a client-side log of keys
 * *this browser* has tested, built entirely from real check responses.
 * It resets on reload; it does not reflect traffic from other clients.
 */
export function useKeyInspector() {
  const [keys, setKeys] = useState<TrackedKey[]>([]);
  const stopFlags = useRef(new Map<string, boolean>());
  const idCounters = useRef(new Map<string, number>());

  function nextId(key: string): number {
    const next = (idCounters.current.get(key) ?? 0) + 1;
    idCounters.current.set(key, next);
    return next;
  }

  function appendResult(key: string, result: SimulatedRequest) {
    setKeys((prev) =>
      prev.map((tracked) =>
        tracked.key === key ? { ...tracked, history: [...tracked.history, result] } : tracked,
      ),
    );
  }

  function addKey(key: string) {
    const trimmed = key.trim();
    if (!trimmed) return;

    setKeys((prev) =>
      prev.some((tracked) => tracked.key === trimmed)
        ? prev
        : [...prev, { key: trimmed, history: [], running: false }],
    );
  }

  function removeKey(key: string) {
    stopFlags.current.delete(key);
    idCounters.current.delete(key);
    setKeys((prev) => prev.filter((tracked) => tracked.key !== key));
  }

  function stop(key: string) {
    stopFlags.current.set(key, true);
  }

  async function run(rawKey: string, count: number) {
    const key = rawKey.trim();
    if (!key) return;

    addKey(key);
    stopFlags.current.set(key, false);
    setKeys((prev) => prev.map((tracked) => (tracked.key === key ? { ...tracked, running: true } : tracked)));

    const total = Math.min(Math.max(Math.floor(count) || 0, 1), MAX_REQUESTS_PER_RUN);

    for (let i = 1; i <= total; i++) {
      if (stopFlags.current.get(key)) break;

      const timestamp = new Date();

      try {
        const response = await checkRateLimit(key);

        appendResult(key, {
          id: nextId(key),
          timestamp,
          allowed: response.allowed,
          remaining: response.remaining,
          retryAfter: response.retry_after,
        });
      } catch (err) {
        appendResult(key, {
          id: nextId(key),
          timestamp,
          allowed: false,
          remaining: 0,
          retryAfter: 0,
          error: err instanceof ApiError ? err.message : "Request failed",
        });

        // Almost certainly a dead API, not a one-off — stop this key's
        // run rather than repeating the same failure `total` times.
        break;
      }

      if (i < total) {
        await new Promise((resolve) => setTimeout(resolve, DELAY_BETWEEN_REQUESTS_MS));
      }
    }

    setKeys((prev) => prev.map((tracked) => (tracked.key === key ? { ...tracked, running: false } : tracked)));
  }

  return { keys, addKey, removeKey, run, stop, maxRequestsPerRun: MAX_REQUESTS_PER_RUN };
}
