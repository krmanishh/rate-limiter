"use client";

import { useEffect, useState } from "react";
import { fetchConfig } from "@/lib/api";
import type { Storage } from "@/lib/types";
import Badge from "@/components/ui/Badge";
import Spinner from "@/components/ui/Spinner";

interface RedisHealthTileProps {
  /** sum(rate(rate_limit_errors_total[...])) — already computed by
   * useObservabilityData, not recomputed here. */
  errorRatePerSecond: number;
}

/** There's no dedicated Redis health check (no redis_exporter in the
 * stack), so this infers health from two things the backend already
 * exposes: which storage backend is active (GET /api/v1/config) and
 * whether rate_limit_errors_total has been incrementing recently. This
 * is a best-effort signal, not a direct probe — labeled as such below. */
export default function RedisHealthTile({ errorRatePerSecond }: RedisHealthTileProps) {
  const [storage, setStorage] = useState<Storage | "unknown">("unknown");
  const [status, setStatus] = useState<"loading" | "ready" | "error">("loading");

  useEffect(() => {
    const controller = new AbortController();

    fetchConfig(controller.signal)
      .then((config) => {
        if (controller.signal.aborted) return;
        setStorage(config.storage);
        setStatus("ready");
      })
      .catch(() => {
        if (controller.signal.aborted) return;
        setStatus("error");
      });

    return () => controller.abort();
  }, []);

  if (status === "loading") {
    return (
      <span className="inline-flex items-center gap-2 text-sm text-slate-500">
        <Spinner /> Checking…
      </span>
    );
  }

  if (status === "error") {
    return <Badge tone="warning">Config unavailable</Badge>;
  }

  if (storage !== "redis") {
    return (
      <span className="flex items-center gap-2">
        <Badge tone="neutral">N/A</Badge>
        <span className="text-xs text-slate-400">
          Active backend is &quot;{storage}&quot;, not Redis.
        </span>
      </span>
    );
  }

  const healthy = errorRatePerSecond === 0;

  return (
    <span className="flex items-center gap-2">
      {healthy ? (
        <Badge tone="success">Healthy</Badge>
      ) : (
        <Badge tone="danger">Degraded</Badge>
      )}
      <span className="text-xs text-slate-400">
        {healthy
          ? "No recent rate limiter errors (inferred)."
          : `${errorRatePerSecond.toFixed(2)} errors/sec (inferred).`}
      </span>
    </span>
  );
}
