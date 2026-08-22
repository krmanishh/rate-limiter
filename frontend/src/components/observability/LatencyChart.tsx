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

/** a = p50, b = p95, both from histogram_quantile() over
 * http_request_duration_seconds — overall HTTP latency (includes the
 * rate-limit check as part of the request), not a limiter-only timer,
 * since no dedicated rate-limiter latency metric exists. */
export default function LatencyChart({ data }: { data: DualTimePoint[] }) {
  if (data.length === 0) {
    return (
      <EmptyState
        title="No latency data yet"
        description="http_request_duration_seconds has no samples in this range."
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
        <YAxis
          tick={{ fontSize: 11 }}
          stroke="currentColor"
          className="text-slate-400"
          width={56}
          tickFormatter={(value: number) => `${(value * 1000).toFixed(0)}ms`}
        />
        <Tooltip
          labelFormatter={(value) => formatTime(Number(value))}
          formatter={(value) => [`${(Number(value) * 1000).toFixed(1)}ms`, undefined]}
          contentStyle={{ fontSize: 12 }}
        />
        <Legend wrapperStyle={{ fontSize: 12 }} />
        <Line type="monotone" dataKey="a" name="p50" stroke="#0ea5e9" strokeWidth={2} dot={false} />
        <Line type="monotone" dataKey="b" name="p95" stroke="#a855f7" strokeWidth={2} dot={false} />
      </LineChart>
    </ResponsiveContainer>
  );
}
