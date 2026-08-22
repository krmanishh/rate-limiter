// Simplified, single-key reimplementations of the five algorithms in
// backend/internal/limiter/{fixedwindow,slidinglog,slidingcounter,
// tokenbucket,leakybucket}/limiter.go, ported line-for-line from their
// Allow() logic so the simulation's behavior actually matches the real
// implementations (not just their high-level description).
//
// This runs entirely in the browser against a synthetic list of
// request offsets — it does not call the backend, use Redis, or read
// live traffic. It exists to let you compare how the five algorithms
// would react to the *same* sequence of requests, which the backend
// itself can't demonstrate (it only ever runs one algorithm at a time,
// chosen at startup via RATE_LIMIT_ALGORITHM).

import type { Algorithm } from "./types";

export interface SimulationParams {
  limit: number;
  windowMs: number;
  capacity: number;
  refillRate: number;
  leakRate: number;
}

export interface SimulatedDecision {
  offsetMs: number;
  allowed: boolean;
  remaining: number;
  retryAfterSeconds: number;
}

function ceilSeconds(ms: number): number {
  if (ms <= 0) return 0;
  return Math.ceil(ms / 1000);
}

// backend/internal/limiter/fixedwindow/limiter.go
function simulateFixedWindow(offsets: number[], { limit, windowMs }: SimulationParams): SimulatedDecision[] {
  let windowStart: number | null = null;
  let count = 0;

  return offsets.map((t) => {
    if (windowStart === null || t - windowStart >= windowMs) {
      windowStart = t;
      count = 1;
      return { offsetMs: t, allowed: true, remaining: limit - 1, retryAfterSeconds: 0 };
    }

    if (count >= limit) {
      return {
        offsetMs: t,
        allowed: false,
        remaining: 0,
        retryAfterSeconds: ceilSeconds(windowMs - (t - windowStart)),
      };
    }

    count++;
    return { offsetMs: t, allowed: true, remaining: limit - count, retryAfterSeconds: 0 };
  });
}

// backend/internal/limiter/slidinglog/limiter.go
function simulateSlidingLog(offsets: number[], { limit, windowMs }: SimulationParams): SimulatedDecision[] {
  let timestamps: number[] = [];

  return offsets.map((t) => {
    const windowStart = t - windowMs;
    timestamps = timestamps.filter((ts) => ts >= windowStart);

    if (timestamps.length >= limit) {
      return {
        offsetMs: t,
        allowed: false,
        remaining: 0,
        retryAfterSeconds: ceilSeconds(timestamps[0] + windowMs - t),
      };
    }

    timestamps.push(t);
    return { offsetMs: t, allowed: true, remaining: limit - timestamps.length, retryAfterSeconds: 0 };
  });
}

// backend/internal/limiter/slidingcounter/limiter.go
function simulateSlidingCounter(offsets: number[], { limit, windowMs }: SimulationParams): SimulatedDecision[] {
  let currentCount = 0;
  let previousCount = 0;
  let windowStart: number | null = null;

  function estimatedCount(elapsed: number): number {
    const previousWeight = (windowMs - elapsed) / windowMs;
    return Math.trunc(previousCount * previousWeight + currentCount);
  }

  function retryAfterMs(elapsed: number): number {
    const remaining = limit - currentCount;

    if (remaining <= 0 || previousCount <= 0) return windowMs - elapsed;

    const fraction = remaining / previousCount;

    if (fraction >= 1) return 0;

    const targetElapsed = windowMs - fraction * windowMs;
    return Math.max(targetElapsed - elapsed, 0);
  }

  function admit(t: number, elapsed: number): SimulatedDecision {
    const estimate = estimatedCount(elapsed);

    if (estimate >= limit) {
      return { offsetMs: t, allowed: false, remaining: 0, retryAfterSeconds: ceilSeconds(retryAfterMs(elapsed)) };
    }

    currentCount++;
    return { offsetMs: t, allowed: true, remaining: Math.max(limit - (estimate + 1), 0), retryAfterSeconds: 0 };
  }

  return offsets.map((t) => {
    if (windowStart === null) {
      windowStart = t;
      currentCount = 1;
      return { offsetMs: t, allowed: true, remaining: limit - 1, retryAfterSeconds: 0 };
    }

    let elapsed = t - windowStart;

    if (elapsed < windowMs) {
      return admit(t, elapsed);
    }

    if (elapsed < 2 * windowMs) {
      previousCount = currentCount;
      currentCount = 0;
      windowStart += windowMs;
    } else {
      previousCount = 0;
      currentCount = 0;
      windowStart = t;
    }

    elapsed = t - windowStart;
    return admit(t, elapsed);
  });
}

// backend/internal/limiter/tokenbucket/limiter.go
function simulateTokenBucket(offsets: number[], { capacity, refillRate }: SimulationParams): SimulatedDecision[] {
  let tokens = capacity;
  let lastRefill: number | null = null;

  return offsets.map((t) => {
    const elapsedSeconds = (t - (lastRefill ?? t)) / 1000;
    tokens = Math.min(capacity, tokens + elapsedSeconds * refillRate);
    lastRefill = t;

    if (tokens < 1) {
      const retryAfterSeconds = (1 - tokens) / refillRate;
      return { offsetMs: t, allowed: false, remaining: 0, retryAfterSeconds: ceilSeconds(retryAfterSeconds * 1000) };
    }

    tokens -= 1;
    return { offsetMs: t, allowed: true, remaining: Math.floor(tokens), retryAfterSeconds: 0 };
  });
}

// backend/internal/limiter/leakybucket/limiter.go
function simulateLeakyBucket(offsets: number[], { capacity, leakRate }: SimulationParams): SimulatedDecision[] {
  let queueSize = 0;
  let lastLeak: number | null = null;

  return offsets.map((t) => {
    const elapsedSeconds = (t - (lastLeak ?? t)) / 1000;
    const leaked = Math.floor(elapsedSeconds * leakRate);

    if (leaked > 0) {
      queueSize = Math.max(0, queueSize - leaked);
      lastLeak = t;
    } else if (lastLeak === null) {
      lastLeak = t;
    }

    if (queueSize >= capacity) {
      const deficit = queueSize - capacity + 1;
      return { offsetMs: t, allowed: false, remaining: 0, retryAfterSeconds: Math.ceil(deficit / leakRate) };
    }

    queueSize++;
    return { offsetMs: t, allowed: true, remaining: capacity - queueSize, retryAfterSeconds: 0 };
  });
}

const SIMULATORS: Record<Algorithm, (offsets: number[], params: SimulationParams) => SimulatedDecision[]> = {
  fixed_window: simulateFixedWindow,
  sliding_log: simulateSlidingLog,
  sliding_counter: simulateSlidingCounter,
  token_bucket: simulateTokenBucket,
  leaky_bucket: simulateLeakyBucket,
};

export function simulateAlgorithm(
  algorithm: Algorithm,
  offsets: number[],
  params: SimulationParams,
): SimulatedDecision[] {
  return SIMULATORS[algorithm](offsets, params);
}

/** Evenly spaced request offsets, in ms, starting at 0 — a small
 * intervalMs models a burst, a larger one models steady traffic. */
export function generateOffsets(count: number, intervalMs: number): number[] {
  return Array.from({ length: count }, (_, i) => i * intervalMs);
}
