"use client";

import { useState } from "react";
import AlgorithmExplorer from "@/components/AlgorithmExplorer";
import ConfigPanel from "@/components/ConfigPanel";
import RequestSimulator from "@/components/RequestSimulator";
import StatusPanel from "@/components/StatusPanel";
import type { ConfigResponse } from "@/lib/types";

export default function Home() {
  const [activeConfig, setActiveConfig] = useState<ConfigResponse | null>(null);

  return (
    <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:px-6 lg:px-8">
      <header className="mb-8">
        <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
          Rate Limiter Playground
        </h1>
        <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
          An interactive dashboard for the Go rate limiter API — fire
          requests at it and watch the limiter respond in real time.
        </p>
      </header>

      <div className="grid gap-6 lg:grid-cols-2">
        <StatusPanel />
        <ConfigPanel onConfigLoaded={setActiveConfig} />
      </div>

      <div className="mt-6">
        <AlgorithmExplorer activeConfig={activeConfig} />
      </div>

      <div className="mt-6">
        <RequestSimulator />
      </div>
    </main>
  );
}
