"use client";

import { useEffect, useRef, useState } from "react";
import { ALGORITHMS, getAlgorithm } from "@/lib/algorithms";
import type { Algorithm, ConfigResponse } from "@/lib/types";
import Badge from "./ui/Badge";
import Card from "./ui/Card";

export default function AlgorithmExplorer({
  activeConfig,
}: {
  activeConfig: ConfigResponse | null;
}) {
  const [selected, setSelected] = useState<Algorithm>("fixed_window");
  const hasSyncedToActive = useRef(false);

  // Sync the selector to whatever the live backend is running, but only
  // once (the first time config loads) — after that, the user is free
  // to explore other algorithms without the selector jumping back.
  useEffect(() => {
    if (activeConfig && !hasSyncedToActive.current) {
      setSelected(activeConfig.algorithm);
      hasSyncedToActive.current = true;
    }
  }, [activeConfig]);

  const meta = getAlgorithm(selected);
  const isActive = activeConfig?.algorithm === selected;

  return (
    <Card
      title="Algorithm explorer"
      description="Compare the five algorithms this API supports. This does not change the live backend."
    >
      <div className="mb-4">
        <label
          htmlFor="algorithm-select"
          className="mb-1 block text-xs font-medium text-slate-500"
        >
          Algorithm
        </label>
        <select
          id="algorithm-select"
          value={selected}
          onChange={(event) => setSelected(event.target.value as Algorithm)}
          className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-700 dark:bg-slate-800 sm:w-72"
        >
          {ALGORITHMS.map((algorithm) => (
            <option key={algorithm.id} value={algorithm.id}>
              {algorithm.name}
            </option>
          ))}
        </select>
      </div>

      {activeConfig && (
        <p className="mb-4 text-xs text-slate-400">
          {isActive ? (
            <>This is the algorithm currently running on the server.</>
          ) : (
            <>
              The server is currently running{" "}
              <strong>{getAlgorithm(activeConfig.algorithm).name}</strong> — use{" "}
              <em>Current configuration</em> above to see its live values.
            </>
          )}
        </p>
      )}

      <p className="mb-4 text-sm text-slate-600 dark:text-slate-300">
        {meta.description}
      </p>

      <div className="mb-5 flex flex-wrap gap-2">
        {meta.fields.map((field) => (
          <Badge key={field.key} tone="neutral">
            {field.label} ({field.unit})
          </Badge>
        ))}
      </div>

      <div className="grid gap-5 sm:grid-cols-2">
        <div>
          <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-emerald-600 dark:text-emerald-400">
            Pros
          </h3>
          <ul className="space-y-1.5 text-sm text-slate-600 dark:text-slate-300">
            {meta.pros.map((pro) => (
              <li key={pro} className="flex gap-2">
                <span aria-hidden className="text-emerald-500">
                  +
                </span>
                {pro}
              </li>
            ))}
          </ul>
        </div>
        <div>
          <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-rose-600 dark:text-rose-400">
            Cons
          </h3>
          <ul className="space-y-1.5 text-sm text-slate-600 dark:text-slate-300">
            {meta.cons.map((con) => (
              <li key={con} className="flex gap-2">
                <span aria-hidden className="text-rose-500">
                  −
                </span>
                {con}
              </li>
            ))}
          </ul>
        </div>
      </div>
    </Card>
  );
}
