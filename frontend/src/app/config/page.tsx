"use client";

import { useState } from "react";
import { useConfig } from "@/hooks/useConfig";
import AlgorithmConfigCard from "@/components/config/AlgorithmConfigCard";
import StorageConfigCard from "@/components/config/StorageConfigCard";
import FailPolicyConfigCard from "@/components/config/FailPolicyConfigCard";
import Button from "@/components/ui/Button";
import ErrorState from "@/components/ui/ErrorState";
import Spinner from "@/components/ui/Spinner";

export default function ConfigPage() {
  const [refreshIndex, setRefreshIndex] = useState(0);
  const state = useConfig(refreshIndex);

  return (
    <main className="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-6 p-4 sm:p-6 lg:p-8">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Configuration</h1>
          <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
            The rate limiter&apos;s active settings, loaded from{" "}
            <code className="font-mono text-xs">GET /api/v1/config</code>.
          </p>
        </div>
        <Button
          variant="secondary"
          onClick={() => setRefreshIndex((n) => n + 1)}
          disabled={state.status === "loading"}
        >
          {state.status === "loading" ? <Spinner /> : "Refresh"}
        </Button>
      </header>

      <div
        role="note"
        className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-900/60 dark:bg-amber-950/40 dark:text-amber-200"
      >
        This is a read-only view. The Go API has no endpoint to change
        rate limit settings at runtime — every value below is set via
        environment variables when the server process starts (see{" "}
        <code className="font-mono text-xs">backend/internal/config/loader.go</code>
        ) and requires a restart to change. The selects below reflect
        this: they show the live value but can&apos;t be edited from here.
      </div>

      <div aria-live="polite">
        {state.status === "loading" && (
          <span className="inline-flex items-center gap-2 text-sm text-slate-500">
            <Spinner /> Loading configuration…
          </span>
        )}

        {state.status === "error" && (
          <ErrorState message={state.message} onRetry={() => setRefreshIndex((n) => n + 1)} />
        )}
      </div>

      {state.status === "ready" && (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <AlgorithmConfigCard config={state.config} />
          <StorageConfigCard config={state.config} />
          <FailPolicyConfigCard config={state.config} />
        </div>
      )}
    </main>
  );
}
