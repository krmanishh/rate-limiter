import type { SimulatedRequest } from "@/lib/types";
import Badge from "./ui/Badge";

export default function ResultsTable({
  results,
  className = "",
}: {
  results: SimulatedRequest[];
  className?: string;
}) {
  if (results.length === 0) return null;

  // Newest first, so the latest result is visible without scrolling.
  const rows = [...results].reverse();

  return (
    <div
      className={`overflow-x-auto rounded-lg border border-slate-200 dark:border-slate-800 ${className}`}
    >
      <table className="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-800">
        <thead className="bg-slate-50 dark:bg-slate-800/60">
          <tr>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">
              #
            </th>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">
              Time
            </th>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">
              Status
            </th>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">
              Remaining
            </th>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">
              Retry-After
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
          {rows.map((result) => (
            <tr key={result.id}>
              <td className="px-3 py-2 text-slate-500 dark:text-slate-400">{result.id}</td>
              <td className="px-3 py-2 font-mono text-xs text-slate-500 dark:text-slate-400">
                {result.timestamp.toLocaleTimeString()}
              </td>
              <td className="px-3 py-2">
                {result.error ? (
                  <Badge tone="danger">Error</Badge>
                ) : result.allowed ? (
                  <Badge tone="success">Allowed</Badge>
                ) : (
                  <Badge tone="danger">Rejected</Badge>
                )}
              </td>
              <td className="px-3 py-2 text-slate-700 dark:text-slate-300">
                {result.error ? "—" : result.remaining}
              </td>
              <td className="px-3 py-2 text-slate-700 dark:text-slate-300">
                {result.error ? "—" : result.retryAfter > 0 ? `${result.retryAfter}s` : "—"}
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {results.some((result) => result.error) && (
        <p className="border-t border-slate-200 px-3 py-2 text-xs text-rose-500 dark:border-slate-800 dark:text-rose-400">
          {results.find((result) => result.error)?.error}
        </p>
      )}
    </div>
  );
}
