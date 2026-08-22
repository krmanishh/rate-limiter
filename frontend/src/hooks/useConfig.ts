"use client";

import { useEffect, useState } from "react";
import { ApiError, fetchConfig } from "@/lib/api";
import type { ConfigResponse } from "@/lib/types";

export type ConfigState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ready"; config: ConfigResponse };

/** Fetches GET /api/v1/config — the server's active rate limit settings,
 * loaded from its environment at startup. Shared by any page that needs
 * the live config rather than each re-implementing the same fetch. */
export function useConfig(refreshIndex = 0): ConfigState {
  const [state, setState] = useState<ConfigState>({ status: "loading" });

  useEffect(() => {
    const controller = new AbortController();

    async function load() {
      setState({ status: "loading" });

      try {
        const config = await fetchConfig(controller.signal);
        if (controller.signal.aborted) return;
        setState({ status: "ready", config });
      } catch (err) {
        if (controller.signal.aborted) return;
        setState({
          status: "error",
          message: err instanceof ApiError ? err.message : "Failed to load configuration",
        });
      }
    }

    load();

    return () => controller.abort();
  }, [refreshIndex]);

  return state;
}
