import type { ConfigResponse } from "@/lib/types";
import Badge from "@/components/ui/Badge";
import Card from "@/components/ui/Card";
import ReadOnlySetting from "./ReadOnlySetting";

const OPTIONS = [
  { value: "memory", label: "In-memory" },
  { value: "redis", label: "Redis" },
];

export default function StorageConfigCard({ config }: { config: ConfigResponse }) {
  return (
    <Card
      title="Storage backend"
      description="Where rate limit counters/state actually live."
    >
      <ReadOnlySetting
        id="config-storage"
        label="Active backend"
        value={config.storage}
        envVar="RATE_LIMIT_STORAGE"
        options={OPTIONS}
      />

      <div className="mt-4 flex items-center gap-2">
        <Badge tone={config.storage === "redis" ? "success" : "neutral"}>
          {config.storage === "redis" ? "Shared across instances" : "Single-process only"}
        </Badge>
      </div>

      <p className="mt-3 text-sm text-slate-600 dark:text-slate-300">
        {config.storage === "redis"
          ? "Counters are stored in Redis, so the limit is enforced consistently across every instance of this API talking to the same Redis."
          : "Counters live in this process's memory. If the API runs as multiple instances (e.g. behind a load balancer), each one enforces its own separate limit — this is fine for local development or a single instance, not for horizontal scaling."}
      </p>
    </Card>
  );
}
