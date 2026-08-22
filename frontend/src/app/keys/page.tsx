"use client";

import { useState } from "react";
import { useKeyInspector } from "@/hooks/useKeyInspector";
import KeyForm from "@/components/keys/KeyForm";
import KeyTable from "@/components/keys/KeyTable";
import KeyAnalyticsPanel from "@/components/keys/KeyAnalyticsPanel";
import Card from "@/components/ui/Card";
import EmptyState from "@/components/ui/EmptyState";

export default function KeysPage() {
  const { keys, run, stop, removeKey, maxRequestsPerRun } = useKeyInspector();
  const [selectedKey, setSelectedKey] = useState<string | null>(null);

  const selected = keys.find((tracked) => tracked.key === selectedKey) ?? null;

  function handleRemove(key: string) {
    removeKey(key);
    if (selectedKey === key) setSelectedKey(null);
  }

  return (
    <main className="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-6 p-4 sm:p-6 lg:p-8">
      <header>
        <h1 className="text-2xl font-semibold">API Keys</h1>
        <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
          Test rate limit keys against the live API and inspect their history.
        </p>
      </header>

      <div
        role="note"
        className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-900/60 dark:bg-amber-950/40 dark:text-amber-200"
      >
        The backend doesn&apos;t maintain a registry of API keys — the{" "}
        <code className="font-mono text-xs">X-API-Key</code> header (see{" "}
        <code className="font-mono text-xs">backend/internal/middleware/ratelimit.go</code>) is
        just an arbitrary string used to partition the rate limit, and there&apos;s no endpoint
        to list or query keys it has seen. Everything below is built from real{" "}
        <code className="font-mono text-xs">POST /api/v1/ratelimit/check</code> responses to keys
        <em> you test here</em>, tracked only in this browser tab for the current session — it
        resets on reload and doesn&apos;t reflect traffic from other clients.
      </div>

      <KeyForm onSubmit={run} maxRequestsPerRun={maxRequestsPerRun} />

      <Card title="Tracked keys" description="Keys tested this session, with their aggregate results.">
        {keys.length === 0 ? (
          <EmptyState
            title="No keys tested yet"
            description="Use the form above to send requests for a key."
          />
        ) : (
          <KeyTable
            keys={keys}
            selectedKey={selectedKey}
            onSelect={setSelectedKey}
            onRun={run}
            onStop={stop}
            onRemove={handleRemove}
            maxRequestsPerRun={maxRequestsPerRun}
          />
        )}
      </Card>

      <KeyAnalyticsPanel tracked={selected} />
    </main>
  );
}
