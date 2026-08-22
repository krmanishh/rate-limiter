import Select from "@/components/ui/Select";

interface Option {
  value: string;
  label: string;
}

interface ReadOnlySettingProps {
  id: string;
  label: string;
  value: string;
  options: Option[];
  envVar: string;
}

/** A disabled <select> pre-filled with the server's actual current
 * value, plus a caption naming the environment variable that controls
 * it. The Go API has no endpoint to change these at runtime (see
 * backend/internal/config/loader.go) — this makes that limitation
 * visible in the UI itself rather than offering a "Save" button that
 * would silently do nothing. */
export default function ReadOnlySetting({ id, label, value, options, envVar }: ReadOnlySettingProps) {
  return (
    <div>
      <Select
        id={id}
        label={label}
        value={value}
        disabled
        onChange={() => {}}
        aria-describedby={`${id}-hint`}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </Select>
      <p id={`${id}-hint`} className="mt-1 text-xs text-slate-400">
        Read-only — set via <code className="font-mono">{envVar}</code> at
        startup; changing it requires restarting the server.
      </p>
    </div>
  );
}
