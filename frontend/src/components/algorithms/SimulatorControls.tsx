export interface SimulatorFormValues {
  limit: number;
  windowSeconds: number;
  capacity: number;
  refillRate: number;
  leakRate: number;
  requestCount: number;
  intervalMs: number;
}

interface Field {
  key: keyof SimulatorFormValues;
  label: string;
  unit: string;
  min: number;
  max: number;
  step?: number;
}

const FIELDS: Field[] = [
  { key: "limit", label: "Limit", unit: "requests", min: 1, max: 100 },
  { key: "windowSeconds", label: "Window", unit: "seconds", min: 1, max: 120 },
  { key: "capacity", label: "Capacity", unit: "tokens/slots", min: 1, max: 100 },
  { key: "refillRate", label: "Refill rate", unit: "tokens/sec", min: 0.1, max: 20, step: 0.1 },
  { key: "leakRate", label: "Leak rate", unit: "req/sec", min: 0.1, max: 20, step: 0.1 },
  { key: "requestCount", label: "# requests to simulate", unit: "requests", min: 1, max: 200 },
  { key: "intervalMs", label: "Interval between requests", unit: "ms", min: 0, max: 5000, step: 10 },
];

interface SimulatorControlsProps {
  values: SimulatorFormValues;
  onChange: (values: SimulatorFormValues) => void;
}

export default function SimulatorControls({ values, onChange }: SimulatorControlsProps) {
  function set(key: keyof SimulatorFormValues, raw: string) {
    const value = Number(raw);
    onChange({ ...values, [key]: Number.isFinite(value) ? value : 0 });
  }

  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-7">
      {FIELDS.map((field) => (
        <div key={field.key}>
          <label htmlFor={`sim-${field.key}`} className="mb-1 block text-xs font-medium text-slate-500">
            {field.label}
          </label>
          <input
            id={`sim-${field.key}`}
            type="number"
            min={field.min}
            max={field.max}
            step={field.step ?? 1}
            value={values[field.key]}
            onChange={(event) => set(field.key, event.target.value)}
            className="w-full rounded-lg border border-slate-300 bg-white px-2 py-1.5 text-sm dark:border-slate-700 dark:bg-slate-800"
          />
          <p className="mt-0.5 text-[11px] text-slate-400">{field.unit}</p>
        </div>
      ))}
    </div>
  );
}
