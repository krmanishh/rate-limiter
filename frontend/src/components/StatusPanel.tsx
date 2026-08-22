"use client";

import { useEffect, useState } from "react";
import { ApiError, API_BASE_URL, fetchHealth } from "@/lib/api";
import Badge from "./ui/Badge";
import Card from "./ui/Card";
import Spinner from "./ui/Spinner";

type Status = "loading" | "online" | "offline";

const POLL_INTERVAL_MS = 10_000;

export default function StatusPanel() {
  const [status, setStatus] = useState<Status>("loading");
  const [error, setError] = useState<string | null>(null);
  const [lastChecked, setLastChecked] = useState<Date | null>(null);

  useEffect(() => {
    const controller = new AbortController();

    async function check() {
      try {
        await fetchHealth(controller.signal);
        if (controller.signal.aborted) return;
        setStatus("online");
        setError(null);
      } catch (err) {
        if (controller.signal.aborted) return;
        setStatus("offline");
        setError(err instanceof ApiError ? err.message : "Unknown error");
      } finally {
        if (!controller.signal.aborted) setLastChecked(new Date());
      }
    }

    check();
    const interval = setInterval(check, POLL_INTERVAL_MS);

    return () => {
      controller.abort();
      clearInterval(interval);
    };
  }, []);

  return (
    <Card title="Server status" description="Polls GET /health every 10s.">
      <div className="flex flex-wrap items-center gap-3">
        {status === "loading" && (
          <span className="inline-flex items-center gap-2 text-sm text-slate-500">
            <Spinner /> Checking…
          </span>
        )}
        {status === "online" && <Badge tone="success">Online</Badge>}
        {status === "offline" && <Badge tone="danger">Offline</Badge>}
        <span className="font-mono text-xs text-slate-400">{API_BASE_URL}</span>
      </div>

      {status === "offline" && error && (
        <p className="mt-2 text-xs text-rose-600 dark:text-rose-400">{error}</p>
      )}

      {lastChecked && (
        <p className="mt-3 text-xs text-slate-400">
          Last checked {lastChecked.toLocaleTimeString()}
        </p>
      )}
    </Card>
  );
}
