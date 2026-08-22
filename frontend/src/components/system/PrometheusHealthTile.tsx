import type { PrometheusSignalState } from "@/hooks/usePrometheusSignal";
import { PROMETHEUS_BASE_URL } from "@/lib/prometheus";
import Badge from "@/components/ui/Badge";
import Spinner from "@/components/ui/Spinner";

/** Prometheus has no CORS-enabled health endpoint (/-/healthy doesn't
 * send Access-Control-Allow-Origin, unlike /api/v1/*), so reachability
 * is inferred from whether a real query against it succeeds. */
export default function PrometheusHealthTile({ signal }: { signal: PrometheusSignalState }) {
  if (signal.status === "loading") {
    return (
      <span className="inline-flex items-center gap-2 text-sm text-slate-500">
        <Spinner /> Checking…
      </span>
    );
  }

  if (signal.status === "unreachable") {
    return (
      <span className="flex flex-col gap-1">
        <Badge tone="danger">Unreachable</Badge>
        <span className="font-mono text-xs text-slate-400">{PROMETHEUS_BASE_URL}</span>
      </span>
    );
  }

  return (
    <span className="flex flex-col gap-1">
      <Badge tone="success">Reachable</Badge>
      <span className="font-mono text-xs text-slate-400">{PROMETHEUS_BASE_URL}</span>
    </span>
  );
}
