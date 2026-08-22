import type { ConfigResponse } from "@/lib/types";
import Badge from "@/components/ui/Badge";
import Card from "@/components/ui/Card";
import ReadOnlySetting from "./ReadOnlySetting";

const OPTIONS = [
  { value: "closed", label: "Fail closed (reject on error)" },
  { value: "open", label: "Fail open (allow on error)" },
];

/** Only meaningful when storage is Redis — a memory limiter has no
 * external dependency to fail (see backend/internal/config/config.go's
 * FailMode doc comment). Shown either way since the server always
 * reports a fail_mode, but labeled "not currently in effect" when it
 * can't apply. */
export default function FailPolicyConfigCard({ config }: { config: ConfigResponse }) {
  const applies = config.storage === "redis";

  return (
    <Card
      title="Redis failure policy"
      description="What happens to a request if the storage backend errors."
    >
      <ReadOnlySetting
        id="config-fail-mode"
        label="Fail mode"
        value={config.fail_mode}
        envVar="RATE_LIMIT_FAIL_MODE"
        options={OPTIONS}
      />

      <div className="mt-4 flex items-center gap-2">
        {applies ? (
          <Badge tone={config.fail_mode === "open" ? "warning" : "success"}>
            In effect (storage is Redis)
          </Badge>
        ) : (
          <Badge tone="neutral">Not currently in effect (storage is memory)</Badge>
        )}
      </div>

      <p className="mt-3 text-sm text-slate-600 dark:text-slate-300">
        {config.fail_mode === "open"
          ? "If Redis becomes unreachable, requests are allowed through unlimited rather than blocked — this prioritizes availability of the protected resource over strict enforcement during an outage."
          : "If Redis becomes unreachable, requests are rejected rather than let through unlimited — this protects the downstream resource at the cost of availability during an outage."}
        {!applies && " An in-memory limiter has no external backend to fail, so this setting has nothing to act on right now."}
      </p>
    </Card>
  );
}
