import type { ReactNode } from "react";

interface MetricTileProps {
  label: string;
  value: ReactNode;
  hint?: string;
}

/** A single KPI number inside a Card — reused for total/allowed/rejected/
 * rejection-rate so the dashboard doesn't need four bespoke components. */
export default function MetricTile({ label, value, hint }: MetricTileProps) {
  return (
    <div className="rounded-lg border border-slate-200 px-4 py-3 dark:border-slate-800">
      <p className="text-xs font-medium uppercase tracking-wide text-slate-500 dark:text-slate-400">
        {label}
      </p>
      <p className="mt-1 text-2xl font-semibold text-slate-900 dark:text-slate-100">
        {value}
      </p>
      {hint && <p className="mt-1 text-xs text-slate-400">{hint}</p>}
    </div>
  );
}
