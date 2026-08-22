// Thin client for Prometheus's HTTP query API:
// https://prometheus.io/docs/prometheus/latest/querying/api/
//
// This talks to Prometheus directly, not the Go backend — the backend
// only exposes a point-in-time /metrics snapshot in exposition format
// (see internal/metrics/metrics.go); actual time-series data (for
// charts) and aggregation (rate(), sum(), histogram_quantile()) only
// exist in Prometheus, which is already scraping and storing them.
// This client queries that existing data — it does not collect,
// compute, or duplicate any metric itself.

export const PROMETHEUS_BASE_URL = (
  process.env.NEXT_PUBLIC_PROMETHEUS_BASE_URL ?? "http://localhost:9090"
).replace(/\/$/, "");

export class PrometheusError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "PrometheusError";
  }
}

/** One label set + its current value, from an instant query. */
export interface VectorSample {
  metric: Record<string, string>;
  /** [unix_seconds, value_as_string] — Prometheus always returns the
   * sample value as a string (it can be "NaN" or "+Inf"). */
  value: [number, string];
}

/** One label set + its values over the queried time range. */
export interface MatrixSeries {
  metric: Record<string, string>;
  values: [number, string][];
}

interface QueryEnvelope<TResultType extends string, TResult> {
  status: "success" | "error";
  errorType?: string;
  error?: string;
  data?: {
    resultType: TResultType;
    result: TResult;
  };
}

type VectorEnvelope = QueryEnvelope<"vector", VectorSample[]>;
type MatrixEnvelope = QueryEnvelope<"matrix", MatrixSeries[]>;

async function get<T extends QueryEnvelope<string, unknown>>(
  path: string,
  params: Record<string, string>,
  signal?: AbortSignal,
): Promise<T> {
  const url = new URL(`${PROMETHEUS_BASE_URL}${path}`);

  for (const [key, value] of Object.entries(params)) {
    url.searchParams.set(key, value);
  }

  let response: Response;

  try {
    response = await fetch(url.toString(), { signal });
  } catch {
    throw new PrometheusError(
      `Could not reach Prometheus at ${PROMETHEUS_BASE_URL}. Is it running?`,
    );
  }

  const body = (await response.json().catch(() => null)) as T | null;

  if (!response.ok || !body || body.status === "error") {
    throw new PrometheusError(
      body?.error ?? `Prometheus query failed with status ${response.status}`,
    );
  }

  return body;
}

/** GET /api/v1/query — the current value of an expression. */
export async function instantQuery(
  query: string,
  signal?: AbortSignal,
): Promise<VectorSample[]> {
  const body = await get<VectorEnvelope>("/api/v1/query", { query }, signal);
  return body.data?.result ?? [];
}

/** GET /api/v1/query_range — an expression evaluated over a time window. */
export async function rangeQuery(
  query: string,
  startSeconds: number,
  endSeconds: number,
  stepSeconds: number,
  signal?: AbortSignal,
): Promise<MatrixSeries[]> {
  const body = await get<MatrixEnvelope>(
    "/api/v1/query_range",
    {
      query,
      start: String(startSeconds),
      end: String(endSeconds),
      step: String(stepSeconds),
    },
    signal,
  );

  return body.data?.result ?? [];
}

/** Reads the (first, and normally only) scalar value out of an instant
 * query's result, defaulting to 0 for "no data yet" (a sparse counter
 * like rate_limit_errors_total that's never incremented simply won't
 * appear in the result set — that's a legitimate "zero", not an error). */
export function firstValue(samples: VectorSample[], fallback = 0): number {
  const raw = samples[0]?.value[1];

  if (raw === undefined) return fallback;

  const parsed = Number(raw);
  return Number.isFinite(parsed) ? parsed : fallback;
}
