"use client";

import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { DualTimePoint } from "@/lib/observabilityTransform";
import EmptyState from "@/components/ui/EmptyState";

function formatTime(unixSeconds: number): string {
  return new Date(unixSeconds * 1000).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });
}

/** a = rate(rate_limit_allowed_total), b = rate(rate_limit_rejected_total) —
 * both requests/sec, overlaid so allow/reject trends are easy to compare. */
export default function AllowedVsRejectedChart({ data }: { data: DualTimePoint[] }) {
  if (data.length === 0) {
    return (
      <EmptyState
        title="No allow/reject data yet"
        description="rate_limit_allowed_total and rate_limit_rejected_total have no samples in this range."
      />
    );
  }

  return (
    <ResponsiveContainer width="100%" height={240}>
      <LineChart data={data} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-slate-200 dark:stroke-slate-800" />
        <XAxis
          dataKey="timestamp"
          tickFormatter={formatTime}
          tick={{ fontSize: 11 }}
          stroke="currentColor"
          className="text-slate-400"
        />
        <YAxis tick={{ fontSize: 11 }} stroke="currentColor" className="text-slate-400" width={48} />
        <Tooltip
          labelFormatter={(value) => formatTime(Number(value))}
          formatter={(value) => [Number(value).toFixed(2), "req/sec"]}
          contentStyle={{ fontSize: 12 }}
        />
        <Legend wrapperStyle={{ fontSize: 12 }} />
        <Line type="monotone" dataKey="a" name="Allowed" stroke="#10b981" strokeWidth={2} dot={false} />
        <Line type="monotone" dataKey="b" name="Rejected" stroke="#f43f5e" strokeWidth={2} dot={false} />
      </LineChart>
    </ResponsiveContainer>
  );
}
