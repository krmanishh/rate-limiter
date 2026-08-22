"use client";

import { useEffect, useState } from "react";
import { PrometheusError, firstValue, instantQuery, rangeQuery } from "@/lib/prometheus";
import {
  algorithmUsageQuery,
  latencyQuantileOverTimeQuery,
  allowedOverTimeQuery,
  redisErrorRateQuery,
  rejectedOverTimeQuery,
  rejectionRatePercentQuery,
  requestsOverTimeQuery,
  totalAllowedQuery,
  totalErrorsQuery,
  totalRejectedQuery,
  totalRequestsQuery,
  type TimeRangePreset,
} from "@/lib/observabilityQueries";
import { toTimeSeries, zipTimeSeries, type DualTimePoint, type TimePoint } from "@/lib/observabilityTransform";

export interface AlgorithmUsage {
  algorithm: string;
  storage: string;
  requests: number;
}

export interface ObservabilityData {
  totals: {
    requests: number;
    allowed: number;
    rejected: number;
    errors: number;
    /** null when there's no traffic yet (division by zero in Prometheus
     * comes back as NaN) — shown as "no data" rather than a misleading 0%. */
    rejectionRatePercent: number | null;
  };
  algorithmUsage: AlgorithmUsage[];
  requestsOverTime: TimePoint[];
  /** a = allowed/sec, b = rejected/sec */
  allowedVsRejectedOverTime: DualTimePoint[];
  /** a = p50 latency (seconds), b = p95 latency (seconds) */
  latencyOverTime: DualTimePoint[];
  redisErrorRatePerSecond: number;
}

export type ObservabilityState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ready"; data: ObservabilityData; lastUpdated: Date };

/**
 * Fetches everything the observability dashboard needs from Prometheus
 * in one batch, re-running whenever the time range changes or
 * refreshIndex is bumped (manual refresh / auto-refresh timer).
 *
 * This queries Prometheus's existing HTTP API — it does not compute,
 * aggregate, or store any metric itself. See lib/observabilityQueries.ts
 * for the exact PromQL, all built from metrics the Go backend already
 * exposes (backend/internal/metrics/metrics.go).
 */
export function useObservabilityData(
  preset: TimeRangePreset,
  refreshIndex: number,
): ObservabilityState {
  const [state, setState] = useState<ObservabilityState>({ status: "loading" });

  useEffect(() => {
    const controller = new AbortController();

    async function load() {
      setState({ status: "loading" });

      const end = Math.floor(Date.now() / 1000);
      const start = end - preset.rangeSeconds;
      const { stepSeconds, rateWindow } = preset;
      const signal = controller.signal;

      try {
        const [
          requestsSamples,
          allowedSamples,
          rejectedSamples,
          errorsSamples,
          rejectionRateSamples,
          algorithmSamples,
          redisErrorRateSamples,
          requestsOverTimeSeries,
          allowedOverTimeSeries,
          rejectedOverTimeSeries,
          p50Series,
          p95Series,
        ] = await Promise.all([
          instantQuery(totalRequestsQuery, signal),
          instantQuery(totalAllowedQuery, signal),
          instantQuery(totalRejectedQuery, signal),
          instantQuery(totalErrorsQuery, signal),
          instantQuery(rejectionRatePercentQuery, signal),
          instantQuery(algorithmUsageQuery, signal),
          instantQuery(redisErrorRateQuery(rateWindow), signal),
          rangeQuery(requestsOverTimeQuery(rateWindow), start, end, stepSeconds, signal),
          rangeQuery(allowedOverTimeQuery(rateWindow), start, end, stepSeconds, signal),
          rangeQuery(rejectedOverTimeQuery(rateWindow), start, end, stepSeconds, signal),
          rangeQuery(latencyQuantileOverTimeQuery(0.5, rateWindow), start, end, stepSeconds, signal),
          rangeQuery(latencyQuantileOverTimeQuery(0.95, rateWindow), start, end, stepSeconds, signal),
        ]);

        if (signal.aborted) return;

        const rejectionRateRaw = rejectionRateSamples[0]?.value[1];
        const rejectionRateParsed = rejectionRateRaw === undefined ? NaN : Number(rejectionRateRaw);

        setState({
          status: "ready",
          data: {
            totals: {
              requests: firstValue(requestsSamples),
              allowed: firstValue(allowedSamples),
              rejected: firstValue(rejectedSamples),
              errors: firstValue(errorsSamples),
              rejectionRatePercent: Number.isFinite(rejectionRateParsed) ? rejectionRateParsed : null,
            },
            algorithmUsage: algorithmSamples.map((sample) => ({
              algorithm: sample.metric.algorithm ?? "unknown",
              storage: sample.metric.storage ?? "unknown",
              requests: Number(sample.value[1]) || 0,
            })),
            requestsOverTime: toTimeSeries(requestsOverTimeSeries),
            allowedVsRejectedOverTime: zipTimeSeries(allowedOverTimeSeries, rejectedOverTimeSeries),
            latencyOverTime: zipTimeSeries(p50Series, p95Series),
            redisErrorRatePerSecond: firstValue(redisErrorRateSamples),
          },
          lastUpdated: new Date(),
        });
      } catch (err) {
        if (signal.aborted) return;

        setState({
          status: "error",
          message: err instanceof PrometheusError ? err.message : "Failed to load metrics",
        });
      }
    }

    load();

    return () => controller.abort();
  }, [preset, refreshIndex]);

  return state;
}
