// Types mirror the Go API's JSON contracts exactly (see
// internal/model/rate_limit.go and internal/api/router.go in the Go
// project). Field names match the wire format (snake_case), not
// TypeScript convention, so there's no risk of silently drifting from
// what the server actually sends.

export type Algorithm =
  | "fixed_window"
  | "sliding_log"
  | "sliding_counter"
  | "token_bucket"
  | "leaky_bucket";

export type Storage = "memory" | "redis";

export type FailMode = "closed" | "open";

/** GET /health */
export interface HealthResponse {
  status: string;
}

/** GET /api/v1/config */
export interface ConfigResponse {
  algorithm: Algorithm;
  storage: Storage;
  fail_mode: FailMode;
  limit: number;
  window_seconds: number;
  capacity: number;
  refill_rate: number;
  leak_rate: number;
}

/** POST /api/v1/ratelimit/check success body (always HTTP 200 — the
 * allowed/rejected decision lives in the body, not the status code). */
export interface RateLimitCheckResponse {
  allowed: boolean;
  remaining: number;
  retry_after: number;
}

/** Error body shape used by every endpoint on a 4xx response. */
export interface ErrorResponse {
  error: string;
}

/** One row of the request simulator's local (client-only) history —
 * not a server type, but kept here alongside the server types since
 * most of its fields come directly from RateLimitCheckResponse. */
export interface SimulatedRequest {
  id: number;
  timestamp: Date;
  allowed: boolean;
  remaining: number;
  retryAfter: number;
  /** Set instead of the above when the request itself failed (network
   * error, non-2xx response) rather than being a normal allow/reject
   * decision. */
  error?: string;
}
