"use client";

import { useState } from "react";
import type { TrackedKey } from "@/hooks/useKeyInspector";
import Badge from "@/components/ui/Badge";
import Button from "@/components/ui/Button";

function summarize(tracked: TrackedKey) {
  const allowed = tracked.history.filter((r) => !r.error && r.allowed).length;
  const rejected = tracked.history.filter((r) => !r.error && !r.allowed).length;
  const errors = tracked.history.filter((r) => r.error).length;
  const last = tracked.history.at(-1);

  return { allowed, rejected, errors, last };
}

interface KeyTableProps {
  keys: TrackedKey[];
  selectedKey: string | null;
  onSelect: (key: string) => void;
  onRun: (key: string, count: number) => void;
  onStop: (key: string) => void;
  onRemove: (key: string) => void;
  maxRequestsPerRun: number;
}

export default function KeyTable({
  keys,
  selectedKey,
  onSelect,
  onRun,
  onStop,
  onRemove,
  maxRequestsPerRun,
}: KeyTableProps) {
  return (
    <div className="overflow-x-auto rounded-lg border border-slate-200 dark:border-slate-800">
      <table className="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-800">
        <thead className="bg-slate-50 dark:bg-slate-800/60">
          <tr>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">Key</th>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">Sent</th>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">Allowed</th>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">Rejected</th>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">Last remaining</th>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">Last checked</th>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">Actions</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
          {keys.map((tracked) => (
            <KeyRow
              key={tracked.key}
              tracked={tracked}
              selected={tracked.key === selectedKey}
              onSelect={onSelect}
              onRun={onRun}
              onStop={onStop}
              onRemove={onRemove}
              maxRequestsPerRun={maxRequestsPerRun}
            />
          ))}
        </tbody>
      </table>
    </div>
  );
}

function KeyRow({
  tracked,
  selected,
  onSelect,
  onRun,
  onStop,
  onRemove,
  maxRequestsPerRun,
}: {
  tracked: TrackedKey;
  selected: boolean;
  onSelect: (key: string) => void;
  onRun: (key: string, count: number) => void;
  onStop: (key: string) => void;
  onRemove: (key: string) => void;
  maxRequestsPerRun: number;
}) {
  const [count, setCount] = useState(5);
  const { allowed, rejected, errors, last } = summarize(tracked);

  return (
    <tr className={selected ? "bg-indigo-50 dark:bg-indigo-500/10" : undefined}>
      <td className="px-3 py-2">
        <button
          type="button"
          onClick={() => onSelect(tracked.key)}
          aria-current={selected ? "true" : undefined}
          className="font-mono text-xs text-indigo-600 hover:underline dark:text-indigo-400"
        >
          {tracked.key}
        </button>
      </td>
      <td className="px-3 py-2 text-slate-600 dark:text-slate-300">{tracked.history.length}</td>
      <td className="px-3 py-2">
        <Badge tone="success">{allowed}</Badge>
      </td>
      <td className="px-3 py-2">
        <Badge tone="danger">{rejected}</Badge>
        {errors > 0 && (
          <span className="ml-1">
            <Badge tone="neutral">{errors} errors</Badge>
          </span>
        )}
      </td>
      <td className="px-3 py-2 text-slate-600 dark:text-slate-300">
        {last && !last.error ? last.remaining : "—"}
      </td>
      <td className="px-3 py-2 font-mono text-xs text-slate-500 dark:text-slate-400">
        {last ? last.timestamp.toLocaleTimeString() : "—"}
      </td>
      <td className="px-3 py-2">
        <div className="flex items-center gap-2">
          <label className="sr-only" htmlFor={`count-${tracked.key}`}>
            Requests to send for {tracked.key}
          </label>
          <input
            id={`count-${tracked.key}`}
            type="number"
            min={1}
            max={maxRequestsPerRun}
            value={count}
            onChange={(event) => setCount(Number(event.target.value))}
            disabled={tracked.running}
            className="w-16 rounded-md border border-slate-300 bg-white px-2 py-1 text-xs disabled:opacity-60 dark:border-slate-700 dark:bg-slate-800"
          />
          {tracked.running ? (
            <Button variant="danger" onClick={() => onStop(tracked.key)} className="px-2 py-1 text-xs">
              Stop
            </Button>
          ) : (
            <Button variant="secondary" onClick={() => onRun(tracked.key, count)} className="px-2 py-1 text-xs">
              Run
            </Button>
          )}
          <Button
            variant="secondary"
            onClick={() => onRemove(tracked.key)}
            disabled={tracked.running}
            aria-label={`Remove ${tracked.key}`}
            className="px-2 py-1 text-xs"
          >
            Remove
          </Button>
        </div>
      </td>
    </tr>
  );
}
