import type { TrackedKey } from "@/hooks/useKeyInspector";
import Card from "@/components/ui/Card";
import EmptyState from "@/components/ui/EmptyState";
import RequestTimeline from "@/components/RequestTimeline";
import ResultsTable from "@/components/ResultsTable";

export default function KeyAnalyticsPanel({ tracked }: { tracked: TrackedKey | null }) {
  if (!tracked) {
    return (
      <Card title="Per-key analytics" description="Select a key from the table to inspect its history.">
        <EmptyState
          title="No key selected"
          description="Test a key above, then click it in the table to see its request history here."
        />
      </Card>
    );
  }

  return (
    <Card
      title={`Per-key analytics — ${tracked.key}`}
      description="This session's real request history for this key, oldest first."
    >
      {tracked.history.length === 0 ? (
        <EmptyState title="No requests sent yet" description="Use Run above to send some." />
      ) : (
        <div className="space-y-4">
          <RequestTimeline results={tracked.history} />
          <ResultsTable results={tracked.history} />
        </div>
      )}
    </Card>
  );
}
