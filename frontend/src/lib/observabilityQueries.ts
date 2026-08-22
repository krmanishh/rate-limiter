// PromQL for every chart/tile in the observability dashboard, built
// only from metrics that actually exist in
// backend/internal/metrics/metrics.go:
//
//   rate_limit_requests_total{algorithm,storage}   counter
//   rate_limit_allowed_total{algorithm,storage}    counter
//   rate_limit_rejected_total{algorithm,storage}   counter
//   rate_limit_errors_total{algorithm,storage}     counter
//   http_requests_total{method,route,status}       counter
//   http_request_duration_seconds{method,route,status}  histogram
//
// There is no dedicated "rate limiter latency" metric — only overall
// HTTP request duration, which includes the rate limit check as part
// of the request but isn't a timer around just that check. The
// latency chart queries http_request_duration_seconds and is labeled
// accordingly rather than implying a metric that doesn't exist.

export interface TimeRangePreset {
  id: string;
  label: string;
  /** Total lookback window, in seconds. */
  rangeSeconds: number;
  /** Step between data points in a range query, in seconds. */
  stepSeconds: number;
  /** The rate()/increase() window used inside queries for this range —
   * wider for longer ranges so the line isn't dominated by scrape-to-
   * scrape noise (scrape_interval is 5s; see backend/prometheus.yml). */
  rateWindow: string;
}

export const TIME_RANGE_PRESETS: TimeRangePreset[] = [
  { id: "5m", label: "Last 5 minutes", rangeSeconds: 5 * 60, stepSeconds: 5, rateWindow: "30s" },
  { id: "15m", label: "Last 15 minutes", rangeSeconds: 15 * 60, stepSeconds: 15, rateWindow: "1m" },
  { id: "1h", label: "Last hour", rangeSeconds: 60 * 60, stepSeconds: 30, rateWindow: "2m" },
  { id: "6h", label: "Last 6 hours", rangeSeconds: 6 * 60 * 60, stepSeconds: 120, rateWindow: "10m" },
];

export const DEFAULT_TIME_RANGE_ID = "15m";

export const REFRESH_INTERVAL_OPTIONS = [
  { id: "off", label: "Off", ms: null },
  { id: "10s", label: "10s", ms: 10_000 },
  { id: "30s", label: "30s", ms: 30_000 },
  { id: "60s", label: "60s", ms: 60_000 },
] as const;

export const DEFAULT_REFRESH_INTERVAL_ID = "30s";

// ---- Instant queries (current totals / breakdowns) ----

export const totalRequestsQuery = "sum(rate_limit_requests_total)";
export const totalAllowedQuery = "sum(rate_limit_allowed_total)";
export const totalRejectedQuery = "sum(rate_limit_rejected_total)";
export const totalErrorsQuery = "sum(rate_limit_errors_total)";

/** Percentage, 0-100. NaN (no requests yet) is handled by the caller,
 * not smoothed over here — a dashboard that quietly shows 0% before
 * any traffic has flowed is more misleading than an explicit "no data". */
export const rejectionRatePercentQuery =
  "100 * sum(rate_limit_rejected_total) / sum(rate_limit_requests_total)";

/** One sample per (algorithm, storage) combination this process has
 * ever run as — in practice one, since RATE_LIMIT_ALGORITHM is fixed
 * per process, but grouped rather than assumed singular. */
export const algorithmUsageQuery = "sum by (algorithm, storage) (rate_limit_requests_total)";

/** Used to infer Redis health: any recent errors mean Redis (or
 * whatever the configured backend is) is failing. This is derived from
 * the existing error counter, not a direct Redis probe — labeled as
 * such in the UI. */
export function redisErrorRateQuery(rateWindow: string): string {
  return `sum(rate(rate_limit_errors_total[${rateWindow}]))`;
}

// ---- Range queries (charts) ----

export function requestsOverTimeQuery(rateWindow: string): string {
  return `sum(rate(rate_limit_requests_total[${rateWindow}]))`;
}

export function allowedOverTimeQuery(rateWindow: string): string {
  return `sum(rate(rate_limit_allowed_total[${rateWindow}]))`;
}

export function rejectedOverTimeQuery(rateWindow: string): string {
  return `sum(rate(rate_limit_rejected_total[${rateWindow}]))`;
}

/** p50/p95 latency in seconds, across all HTTP routes. Uses the
 * existing http_request_duration_seconds histogram — see the module
 * doc comment above for why this isn't a "rate limiter only" number. */
export function latencyQuantileOverTimeQuery(quantile: number, rateWindow: string): string {
  return `histogram_quantile(${quantile}, sum(rate(http_request_duration_seconds_bucket[${rateWindow}])) by (le))`;
}
