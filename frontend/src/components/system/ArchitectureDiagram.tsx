import type { ConfigResponse } from "@/lib/types";
import Badge from "@/components/ui/Badge";

function FlowNode({
  title,
  detail,
  highlight,
}: {
  title: string;
  detail?: string;
  highlight?: boolean;
}) {
  return (
    <div
      className={`min-w-[10rem] flex-1 rounded-lg border px-3 py-2 text-center ${
        highlight
          ? "border-indigo-300 bg-indigo-50 dark:border-indigo-700 dark:bg-indigo-500/10"
          : "border-slate-200 bg-slate-50 dark:border-slate-800 dark:bg-slate-800/40"
      }`}
    >
      <p className="text-sm font-medium text-slate-800 dark:text-slate-100">{title}</p>
      {detail && <p className="mt-0.5 text-xs text-slate-500 dark:text-slate-400">{detail}</p>}
    </div>
  );
}

function Arrow() {
  return (
    <span aria-hidden className="flex shrink-0 items-center justify-center text-slate-300 dark:text-slate-600">
      <span className="hidden sm:inline">&rarr;</span>
      <span className="sm:hidden">&darr;</span>
    </span>
  );
}

/** A static picture of the request path and observability path,
 * assembled from what backend/cmd/server/main.go actually wires up
 * (CORS -> router -> rate limit middleware -> limiter -> store), with
 * the live algorithm/storage from GET /api/v1/config annotated on top
 * where available. Plain HTML/CSS — no diagramming library. */
export default function ArchitectureDiagram({ config }: { config: ConfigResponse | null }) {
  return (
    <div className="flex flex-col gap-6">
      <div>
        <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-400">
          Request path
        </h3>
        <div className="flex flex-col items-stretch gap-2 sm:flex-row sm:items-center">
          <FlowNode title="Browser / client" detail="Sends X-API-Key or connects by IP" />
          <Arrow />
          <FlowNode title="Go API" detail="CORS + HTTP metrics middleware" />
          <Arrow />
          <FlowNode
            title="Rate limit middleware"
            detail={config ? `Algorithm: ${config.algorithm}` : "Algorithm: —"}
            highlight
          />
          <Arrow />
          <FlowNode
            title={config?.storage === "redis" ? "Redis" : "In-memory store"}
            detail={config ? `Storage: ${config.storage}` : "Storage: —"}
            highlight
          />
        </div>
      </div>

      <div>
        <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-400">
          Observability path
        </h3>
        <div className="flex flex-col items-stretch gap-2 sm:flex-row sm:items-center">
          <FlowNode title="Go API" detail="Exposes GET /metrics" />
          <Arrow />
          <FlowNode title="Prometheus" detail="Scrapes /metrics every 5s" />
          <Arrow />
          <FlowNode title="This dashboard" detail="Queries Prometheus's HTTP API directly" />
        </div>
      </div>

      {config && (
        <div className="flex flex-wrap gap-2 border-t border-slate-200 pt-4 dark:border-slate-800">
          <Badge tone="info">algorithm: {config.algorithm}</Badge>
          <Badge tone={config.storage === "redis" ? "success" : "neutral"}>
            storage: {config.storage}
          </Badge>
          <Badge tone={config.fail_mode === "open" ? "warning" : "neutral"}>
            fail-{config.fail_mode}
          </Badge>
        </div>
      )}
    </div>
  );
}
