import type { MatrixSeries } from "./prometheus";

export interface TimePoint {
  timestamp: number; // unix seconds
  value: number;
}

/** Prometheus range queries return one series per distinct label set;
 * our queries are all wrapped in sum(...) so there's exactly one
 * (label-less) series when there's any data at all. */
export function toTimeSeries(series: MatrixSeries[]): TimePoint[] {
  const first = series[0];

  if (!first) return [];

  return first.values.map(([timestamp, value]) => ({
    timestamp,
    value: Number(value) || 0,
  }));
}

export interface DualTimePoint {
  timestamp: number;
  a: number;
  b: number;
}

/** Merges two range queries run with identical start/end/step into one
 * timestamp-aligned array, for charts that overlay two series (e.g.
 * allowed vs rejected, p50 vs p95). */
export function zipTimeSeries(
  seriesA: MatrixSeries[],
  seriesB: MatrixSeries[],
): DualTimePoint[] {
  const a = toTimeSeries(seriesA);
  const b = toTimeSeries(seriesB);
  const length = Math.max(a.length, b.length);
  const points: DualTimePoint[] = [];

  for (let i = 0; i < length; i++) {
    const timestamp = a[i]?.timestamp ?? b[i]?.timestamp;

    if (timestamp === undefined) continue;

    points.push({
      timestamp,
      a: a[i]?.value ?? 0,
      b: b[i]?.value ?? 0,
    });
  }

  return points;
}
