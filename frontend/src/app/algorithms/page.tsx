"use client";

import { useMemo, useState } from "react";
import Card from "@/components/ui/Card";
import ComparisonTable from "@/components/algorithms/ComparisonTable";
import SimulatorControls, { type SimulatorFormValues } from "@/components/algorithms/SimulatorControls";
import SimulatorResults from "@/components/algorithms/SimulatorResults";

const DEFAULT_VALUES: SimulatorFormValues = {
  limit: 5,
  windowSeconds: 10,
  capacity: 10,
  refillRate: 2,
  leakRate: 2,
  requestCount: 20,
  intervalMs: 300,
};

export default function AlgorithmsPage() {
  const [values, setValues] = useState<SimulatorFormValues>(DEFAULT_VALUES);

  const params = useMemo(
    () => ({
      limit: values.limit,
      windowMs: values.windowSeconds * 1000,
      capacity: values.capacity,
      refillRate: values.refillRate,
      leakRate: values.leakRate,
    }),
    [values],
  );

  return (
    <main className="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-6 p-4 sm:p-6 lg:p-8">
      <header>
        <h1 className="text-2xl font-semibold">Algorithms</h1>
        <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
          Compare the five rate limiting algorithms this API supports, and simulate how each
          would react to the same request pattern.
        </p>
      </header>

      <Card title="Comparison" description="Static reference — does not depend on the live server.">
        <ComparisonTable />
      </Card>

      <Card
        title="Interactive simulator"
        description="Runs simplified versions of each algorithm's decision logic in your browser against a synthetic request pattern."
      >
        <div
          role="note"
          className="mb-4 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-900/60 dark:bg-amber-950/40 dark:text-amber-200"
        >
          This simulation runs entirely in your browser — it does not call the backend or use
          Redis. The logic is ported from each algorithm&apos;s real Go implementation in{" "}
          <code className="font-mono text-xs">backend/internal/limiter/</code>, but the server
          only ever runs <em>one</em> algorithm at a time (chosen at startup), so this is the only
          way to see all five react to an identical sequence of requests side by side.
        </div>

        <SimulatorControls values={values} onChange={setValues} />

        <div className="mt-6">
          <SimulatorResults
            params={params}
            requestCount={Math.min(Math.max(Math.floor(values.requestCount) || 0, 1), 200)}
            intervalMs={Math.max(values.intervalMs, 0)}
          />
        </div>
      </Card>
    </main>
  );
}
