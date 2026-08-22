import {
  REFRESH_INTERVAL_OPTIONS,
  TIME_RANGE_PRESETS,
} from "@/lib/observabilityQueries";
import Button from "@/components/ui/Button";
import Select from "@/components/ui/Select";

interface TimeRangeControlProps {
  timeRangeId: string;
  onTimeRangeChange: (id: string) => void;
  refreshIntervalId: string;
  onRefreshIntervalChange: (id: string) => void;
  onRefreshNow: () => void;
  lastUpdated: Date | null;
}

export default function TimeRangeControl({
  timeRangeId,
  onTimeRangeChange,
  refreshIntervalId,
  onRefreshIntervalChange,
  onRefreshNow,
  lastUpdated,
}: TimeRangeControlProps) {
  return (
    <div className="flex flex-wrap items-end gap-3">
      <Select
        label="Time range"
        id="time-range"
        value={timeRangeId}
        onChange={(e) => onTimeRangeChange(e.target.value)}
      >
        {TIME_RANGE_PRESETS.map((preset) => (
          <option key={preset.id} value={preset.id}>
            {preset.label}
          </option>
        ))}
      </Select>

      <Select
        label="Auto-refresh"
        id="refresh-interval"
        value={refreshIntervalId}
        onChange={(e) => onRefreshIntervalChange(e.target.value)}
      >
        {REFRESH_INTERVAL_OPTIONS.map((option) => (
          <option key={option.id} value={option.id}>
            {option.label}
          </option>
        ))}
      </Select>

      <Button variant="secondary" onClick={onRefreshNow}>
        Refresh now
      </Button>

      {lastUpdated && (
        <span className="pb-2 text-xs text-slate-400">
          Last updated {lastUpdated.toLocaleTimeString()}
        </span>
      )}
    </div>
  );
}
