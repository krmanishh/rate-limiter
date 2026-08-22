import type { SimulatedRequest } from "@/lib/types";

function outcomeOf(result: SimulatedRequest): "allowed" | "rejected" | "error" {
  if (result.error) return "error";
  return result.allowed ? "allowed" : "rejected";
}

const OUTCOME_COLOR: Record<ReturnType<typeof outcomeOf>, string> = {
  allowed: "bg-emerald-500",
  rejected: "bg-rose-500",
  error: "bg-slate-400",
};

export default function RequestTimeline({
  results,
  formatTime = (result) => result.timestamp.toLocaleTimeString(),
}: {
  results: SimulatedRequest[];
  /** Defaults to a wall-clock time, which is what real request results
   * carry. The algorithm simulator passes synthetic ms-offset
   * timestamps instead, which need a different label to stay useful
   * (they'd otherwise all show the same clock-time). */
  formatTime?: (result: SimulatedRequest) => string;
}) {
  if (results.length === 0) return null;

  const allowedCount = results.filter((r) => outcomeOf(r) === "allowed").length;
  const rejectedCount = results.filter((r) => outcomeOf(r) === "rejected").length;
  const errorCount = results.filter((r) => outcomeOf(r) === "error").length;

  return (
    <div>
      <div className="mb-2 flex flex-wrap items-center gap-4 text-xs text-slate-500 dark:text-slate-400">
        <Legend color={OUTCOME_COLOR.allowed} label={`${allowedCount} allowed`} />
        <Legend color={OUTCOME_COLOR.rejected} label={`${rejectedCount} rejected`} />
        {errorCount > 0 && (
          <Legend color={OUTCOME_COLOR.error} label={`${errorCount} errors`} />
        )}
      </div>

      <div
        role="list"
        aria-label="Request timeline, oldest first"
        className="flex flex-wrap gap-1"
      >
        {results.map((result) => {
          const outcome = outcomeOf(result);
          const label = `Request ${result.id} at ${formatTime(result)}: ${
            result.error ?? (result.allowed ? "allowed" : "rejected")
          }${!result.error ? `, ${result.remaining} remaining` : ""}`;

          return (
            <div
              key={result.id}
              role="listitem"
              title={label}
              aria-label={label}
              className={`h-6 w-3 rounded-sm ${OUTCOME_COLOR[outcome]}`}
            />
          );
        })}
      </div>
    </div>
  );
}

function Legend({ color, label }: { color: string; label: string }) {
  return (
    <span className="inline-flex items-center gap-1.5">
      <span aria-hidden className={`inline-block h-2 w-2 rounded-full ${color}`} />
      {label}
    </span>
  );
}
