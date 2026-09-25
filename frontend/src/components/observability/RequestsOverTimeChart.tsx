"use client";

import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { TimePoint } from "@/lib/observabilityTransform";
import EmptyState from "@/components/ui/EmptyState";

function formatTime(unixSeconds: number): string {
  return new Date(unixSeconds * 1000).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });
}

/** sum(rate(rate_limit_requests_total[...])) over the selected time range —
 * requests/sec, not a raw count, since the underlying metric is a counter. */
export default function RequestsOverTimeChart({ data }: { data: TimePoint[] }) {
  if (data.length === 0) {
    return (
      <EmptyState
        title="No request data yet"
        description="rate_limit_requests_total has no samples in this range."
      />
    );
  }

  return (
    <ResponsiveContainer width="100%" height={240}>
      <AreaChart data={data} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
        <defs>
          <linearGradient id="requestsFill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#2563eb" stopOpacity={0.35} />
            <stop offset="100%" stopColor="#2563eb" stopOpacity={0} />
          </linearGradient>
        </defs>
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
        <Area
          type="monotone"
          dataKey="value"
          name="req/sec"
          stroke="#2563eb"
          fill="url(#requestsFill)"
          strokeWidth={2}
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}
