"use client";

import { useEffect, useState } from "react";
import { PrometheusError, firstValue, instantQuery } from "@/lib/prometheus";
import { redisErrorRateQuery } from "@/lib/observabilityQueries";

export type PrometheusSignalState =
  | { status: "loading" }
  | { status: "unreachable"; message: string }
  | { status: "ready"; redisErrorRatePerSecond: number };

/**
 * A single, cheap Prometheus query used two ways on the System page:
 * whether the query succeeds at all is Prometheus's own reachability
 * signal (there's no CORS-enabled Prometheus health endpoint to hit
 * directly — see lib/prometheus.ts), and its value feeds the same
 * Redis-health inference RedisHealthTile already uses on the
 * Observability page. One query, not a duplicate metrics pipeline.
 */
export function usePrometheusSignal(refreshIndex = 0): PrometheusSignalState {
  const [state, setState] = useState<PrometheusSignalState>({ status: "loading" });

  useEffect(() => {
    const controller = new AbortController();

    async function load() {
      setState({ status: "loading" });

      try {
        const samples = await instantQuery(redisErrorRateQuery("1m"), controller.signal);
        if (controller.signal.aborted) return;
        setState({ status: "ready", redisErrorRatePerSecond: firstValue(samples) });
      } catch (err) {
        if (controller.signal.aborted) return;
        setState({
          status: "unreachable",
          message: err instanceof PrometheusError ? err.message : "Failed to reach Prometheus",
        });
      }
    }

    load();

    return () => controller.abort();
  }, [refreshIndex]);

  return state;
}
