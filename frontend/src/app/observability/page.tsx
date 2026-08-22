"use client";

import { useEffect, useState } from "react";
import { useObservabilityData } from "@/hooks/useObservabilityData";
import {
  DEFAULT_REFRESH_INTERVAL_ID,
  DEFAULT_TIME_RANGE_ID,
  REFRESH_INTERVAL_OPTIONS,
  TIME_RANGE_PRESETS,
} from "@/lib/observabilityQueries";
import Card from "@/components/ui/Card";
import ErrorState from "@/components/ui/ErrorState";
import Spinner from "@/components/ui/Spinner";
import StatusPanel from "@/components/StatusPanel";
import MetricTile from "@/components/observability/MetricTile";
import TimeRangeControl from "@/components/observability/TimeRangeControl";
import RequestsOverTimeChart from "@/components/observability/RequestsOverTimeChart";
import AllowedVsRejectedChart from "@/components/observability/AllowedVsRejectedChart";
import AlgorithmUsageChart from "@/components/observability/AlgorithmUsageChart";
import LatencyChart from "@/components/observability/LatencyChart";
import RedisHealthTile from "@/components/observability/RedisHealthTile";

/** Ticks `onTick` every `ms` milliseconds (or never, if `ms` is null —
 * the "Off" refresh option). Kept local to this page since scheduling
 * is a UI concern, not something useObservabilityData needs to know. */
function useInterval(ms: number | null, onTick: () => void) {
  useEffect(() => {
    if (ms === null) return;
    const id = setInterval(onTick, ms);
    return () => clearInterval(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps -- onTick intentionally excluded so the timer isn't reset every render
  }, [ms]);
}

export default function ObservabilityPage() {
  const [timeRangeId, setTimeRangeId] = useState(DEFAULT_TIME_RANGE_ID);
  const [refreshIntervalId, setRefreshIntervalId] = useState(DEFAULT_REFRESH_INTERVAL_ID);
  const [refreshIndex, setRefreshIndex] = useState(0);

  const preset =
    TIME_RANGE_PRESETS.find((p) => p.id === timeRangeId) ?? TIME_RANGE_PRESETS[0];

  const bump = () => setRefreshIndex((n) => n + 1);
  const refreshMs =
    REFRESH_INTERVAL_OPTIONS.find((o) => o.id === refreshIntervalId)?.ms ?? null;
  useInterval(refreshMs, bump);

  const state = useObservabilityData(preset, refreshIndex);

  return (
    <main className="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-6 p-6">
      <header>
        <h1 className="text-2xl font-semibold">Observability Dashboard</h1>
        <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
          Live metrics from Prometheus, scraping the Go rate limiter&apos;s{" "}
          <code className="font-mono text-xs">/metrics</code> endpoint.
        </p>
      </header>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <StatusPanel />
        <Card title="Redis health" description="Inferred from rate_limit_errors_total.">
          {state.status === "ready" ? (
            <RedisHealthTile errorRatePerSecond={state.data.redisErrorRatePerSecond} />
          ) : (
            <span className="inline-flex items-center gap-2 text-sm text-slate-500">
              <Spinner /> Waiting for metrics…
            </span>
          )}
        </Card>
      </div>

      <Card>
        <TimeRangeControl
          timeRangeId={timeRangeId}
          onTimeRangeChange={setTimeRangeId}
          refreshIntervalId={refreshIntervalId}
          onRefreshIntervalChange={setRefreshIntervalId}
          onRefreshNow={bump}
          lastUpdated={state.status === "ready" ? state.lastUpdated : null}
        />
      </Card>

      {state.status === "error" && (
        <Card title="Prometheus unavailable">
          <ErrorState message={state.message} onRetry={bump} />
        </Card>
      )}

      {state.status === "loading" && (
        <Card>
          <span className="inline-flex items-center gap-2 text-sm text-slate-500">
            <Spinner /> Loading metrics…
          </span>
        </Card>
      )}

      {state.status === "ready" && (
        <>
          <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
            <MetricTile label="Total requests" value={state.data.totals.requests.toLocaleString()} />
            <MetricTile label="Allowed" value={state.data.totals.allowed.toLocaleString()} />
            <MetricTile label="Rejected" value={state.data.totals.rejected.toLocaleString()} />
            <MetricTile
              label="Rejection rate"
              value={
                state.data.totals.rejectionRatePercent === null
                  ? "No data"
                  : `${state.data.totals.rejectionRatePercent.toFixed(1)}%`
              }
            />
          </div>

          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <Card title="Requests over time" description="sum(rate(rate_limit_requests_total))">
              <RequestsOverTimeChart data={state.data.requestsOverTime} />
            </Card>
            <Card title="Allowed vs rejected" description="req/sec, side by side">
              <AllowedVsRejectedChart data={state.data.allowedVsRejectedOverTime} />
            </Card>
            <Card title="Algorithm usage" description="sum by (algorithm, storage)">
              <AlgorithmUsageChart usage={state.data.algorithmUsage} />
            </Card>
            <Card
              title="HTTP latency (p50 / p95)"
              description="http_request_duration_seconds — overall request latency, not a limiter-only timer"
            >
              <LatencyChart data={state.data.latencyOverTime} />
            </Card>
          </div>
        </>
      )}
    </main>
  );
}
