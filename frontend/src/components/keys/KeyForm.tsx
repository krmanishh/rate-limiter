"use client";

import { useState, type FormEvent } from "react";
import Button from "@/components/ui/Button";
import Card from "@/components/ui/Card";

interface KeyFormProps {
  onSubmit: (key: string, count: number) => void;
  maxRequestsPerRun: number;
}

export default function KeyForm({ onSubmit, maxRequestsPerRun }: KeyFormProps) {
  const [key, setKey] = useState("");
  const [count, setCount] = useState(10);

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    if (!key.trim()) return;
    onSubmit(key, count);
  }

  return (
    <Card
      title="Test a key"
      description="Sends real requests to POST /api/v1/ratelimit/check with this key in X-API-Key."
    >
      <form onSubmit={handleSubmit} className="flex flex-col gap-3 sm:flex-row sm:items-end">
        <div className="flex-1">
          <label htmlFor="key-form-key" className="mb-1 block text-xs font-medium text-slate-500">
            API key / client identifier
          </label>
          <input
            id="key-form-key"
            value={key}
            onChange={(event) => setKey(event.target.value)}
            placeholder="e.g. mobile-app-prod"
            className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-700 dark:bg-slate-800"
          />
        </div>

        <div className="w-full sm:w-40">
          <label htmlFor="key-form-count" className="mb-1 block text-xs font-medium text-slate-500">
            # requests (max {maxRequestsPerRun})
          </label>
          <input
            id="key-form-count"
            type="number"
            min={1}
            max={maxRequestsPerRun}
            value={count}
            onChange={(event) => setCount(Number(event.target.value))}
            className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-700 dark:bg-slate-800"
          />
        </div>

        <Button type="submit" disabled={!key.trim()}>
          Run
        </Button>
      </form>
    </Card>
  );
}
