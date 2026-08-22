"use client";

import { useEffect, useState } from "react";
import { ApiError, fetchConfig } from "@/lib/api";
import { fieldValue, getAlgorithm } from "@/lib/algorithms";
import type { ConfigResponse } from "@/lib/types";
import Badge from "./ui/Badge";
import Button from "./ui/Button";
import Card from "./ui/Card";
import ErrorState from "./ui/ErrorState";
import Spinner from "./ui/Spinner";

type State =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ready"; config: ConfigResponse };

export default function ConfigPanel({
  onConfigLoaded,
}: {
  onConfigLoaded?: (config: ConfigResponse) => void;
}) {
  const [state, setState] = useState<State>({ status: "loading" });
  // Bumping this re-runs the effect below, which is how the "Refresh"
  // button re-fetches without needing a function defined outside the
  // effect (calling an outer async helper from inside useEffect trips
  // react-hooks/set-state-in-effect, since the linter can't see past
  // the closure that the setState calls only happen after an await).
  const [refreshIndex, setRefreshIndex] = useState(0);

  useEffect(() => {
    const controller = new AbortController();

    async function load() {
      setState({ status: "loading" });

      try {
        const config = await fetchConfig(controller.signal);
        if (controller.signal.aborted) return;
        setState({ status: "ready", config });
        onConfigLoaded?.(config);
      } catch (err) {
        if (controller.signal.aborted) return;
        setState({
          status: "error",
          message:
            err instanceof ApiError
              ? err.message
              : "Failed to load configuration",
        });
      }
    }

    load();

    return () => controller.abort();
  }, [refreshIndex, onConfigLoaded]);

  return (
    <Card
      title="Current configuration"
      description="Loaded from the server's environment at startup."
      action={
        <Button
          variant="secondary"
          onClick={() => setRefreshIndex((n) => n + 1)}
          disabled={state.status === "loading"}
        >
          {state.status === "loading" ? <Spinner /> : "Refresh"}
        </Button>
      }
    >
      {state.status === "loading" && (
        <span className="inline-flex items-center gap-2 text-sm text-slate-500">
          <Spinner /> Loading configuration…
        </span>
      )}

      {state.status === "error" && (
        <ErrorState
          message={state.message}
          onRetry={() => setRefreshIndex((n) => n + 1)}
        />
      )}

      {state.status === "ready" && (
        <div className="space-y-4">
          <div className="flex flex-wrap items-center gap-2">
            <Badge tone="info">{getAlgorithm(state.config.algorithm).name}</Badge>
            <Badge tone={state.config.storage === "redis" ? "success" : "neutral"}>
              storage: {state.config.storage}
            </Badge>
            <Badge tone={state.config.fail_mode === "open" ? "warning" : "neutral"}>
              fail-{state.config.fail_mode}
            </Badge>
          </div>

          <dl className="grid grid-cols-2 gap-3 sm:grid-cols-3">
            {getAlgorithm(state.config.algorithm).fields.map((field) => (
              <div
                key={field.key}
                className="rounded-lg bg-slate-100 p-3 dark:bg-slate-800/60"
              >
                <dt className="text-xs uppercase tracking-wide text-slate-400">
                  {field.label}
                </dt>
                <dd className="mt-1 text-lg font-semibold text-slate-800 dark:text-slate-100">
                  {fieldValue(field.key, state.config)}{" "}
                  <span className="text-xs font-normal text-slate-400">
                    {field.unit}
                  </span>
                </dd>
              </div>
            ))}
          </dl>
        </div>
      )}
    </Card>
  );
}
