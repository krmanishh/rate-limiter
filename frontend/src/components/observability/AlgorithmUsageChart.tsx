import type { AlgorithmUsage } from "@/hooks/useObservabilityData";
import Badge from "@/components/ui/Badge";
import EmptyState from "@/components/ui/EmptyState";

/** sum by (algorithm, storage) (rate_limit_requests_total) — normally a
 * single bar, since RATE_LIMIT_ALGORITHM is fixed per running process,
 * but rendered as a list so multiple combinations (e.g. across restarts
 * with a different algorithm) show up correctly. */
export default function AlgorithmUsageChart({ usage }: { usage: AlgorithmUsage[] }) {
  if (usage.length === 0) {
    return (
      <EmptyState
        title="No algorithm usage yet"
        description="rate_limit_requests_total has no samples in this range."
      />
    );
  }

  const max = Math.max(...usage.map((u) => u.requests), 1);
  const sorted = [...usage].sort((a, b) => b.requests - a.requests);

  return (
    <div className="flex flex-col gap-3">
      {sorted.map((entry) => (
        <div key={`${entry.algorithm}-${entry.storage}`}>
          <div className="mb-1 flex items-center justify-between gap-2 text-sm">
            <span className="flex items-center gap-2 font-medium text-slate-700 dark:text-slate-200">
              {entry.algorithm}
              <Badge tone="info">{entry.storage}</Badge>
            </span>
            <span className="font-mono text-xs text-slate-500">
              {entry.requests.toLocaleString()}
            </span>
          </div>
          <div className="h-2 w-full overflow-hidden rounded-full bg-slate-100 dark:bg-slate-800">
            <div
              className="h-full rounded-full bg-indigo-500"
              style={{ width: `${(entry.requests / max) * 100}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  );
}
