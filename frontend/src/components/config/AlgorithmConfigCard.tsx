import { ALGORITHMS, fieldValue, getAlgorithm } from "@/lib/algorithms";
import type { ConfigResponse } from "@/lib/types";
import Card from "@/components/ui/Card";
import ReadOnlySetting from "./ReadOnlySetting";

export default function AlgorithmConfigCard({ config }: { config: ConfigResponse }) {
  const meta = getAlgorithm(config.algorithm);

  return (
    <Card title="Algorithm" description={meta.summary}>
      <ReadOnlySetting
        id="config-algorithm"
        label="Active algorithm"
        value={config.algorithm}
        envVar="RATE_LIMIT_ALGORITHM"
        options={ALGORITHMS.map((algorithm) => ({ value: algorithm.id, label: algorithm.name }))}
      />

      <dl className="mt-4 grid grid-cols-2 gap-3">
        {meta.fields.map((field) => (
          <div key={field.key} className="rounded-lg bg-slate-100 p-3 dark:bg-slate-800/60">
            <dt className="text-xs uppercase tracking-wide text-slate-400">{field.label}</dt>
            <dd className="mt-1 text-lg font-semibold text-slate-800 dark:text-slate-100">
              {fieldValue(field.key, config)}{" "}
              <span className="text-xs font-normal text-slate-400">{field.unit}</span>
            </dd>
          </div>
        ))}
      </dl>
    </Card>
  );
}
