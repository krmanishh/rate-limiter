"use client";

import { useState } from "react";
import { useConfig } from "@/hooks/useConfig";
import { usePrometheusSignal } from "@/hooks/usePrometheusSignal";
import Card from "@/components/ui/Card";
import Button from "@/components/ui/Button";
import Spinner from "@/components/ui/Spinner";
import StatusPanel from "@/components/StatusPanel";
import ArchitectureDiagram from "@/components/system/ArchitectureDiagram";
import PrometheusHealthTile from "@/components/system/PrometheusHealthTile";
import RedisHealthTile from "@/components/observability/RedisHealthTile";
import Badge from "@/components/ui/Badge";

export default function SystemPage() {
  const [refreshIndex, setRefreshIndex] = useState(0);
  const configState = useConfig(refreshIndex);
  const prometheusSignal = usePrometheusSignal(refreshIndex);

  return (
    <main className="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-6 p-4 sm:p-6 lg:p-8">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">System</h1>
          <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
            How the pieces fit together, and whether each one is currently reachable.
          </p>
        </div>
        <Button
          variant="secondary"
          onClick={() => setRefreshIndex((n) => n + 1)}
          disabled={configState.status === "loading"}
        >
          {configState.status === "loading" ? <Spinner /> : "Refresh"}
        </Button>
      </header>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3" aria-live="polite">
        <Card title="Backend">
          <StatusPanel />
        </Card>
        <Card title="Prometheus" description="Reachability inferred from a live query.">
          <PrometheusHealthTile signal={prometheusSignal} />
        </Card>
        <Card title="Redis" description="Inferred from rate_limit_errors_total.">
          {configState.status !== "ready" ? (
            <span className="inline-flex items-center gap-2 text-sm text-slate-500">
              <Spinner /> Waiting for configuration…
            </span>
          ) : prometheusSignal.status === "ready" ? (
            <RedisHealthTile errorRatePerSecond={prometheusSignal.redisErrorRatePerSecond} />
          ) : (
            <span className="flex items-center gap-2">
              <Badge tone="neutral">Unavailable</Badge>
              <span className="text-xs text-slate-400">Prometheus is unreachable.</span>
            </span>
          )}
        </Card>
      </div>

      <Card
        title="Architecture"
        description="The request path and observability path this API is actually wired up as."
      >
        <ArchitectureDiagram config={configState.status === "ready" ? configState.config : null} />
      </Card>
    </main>
  );
}
