import { ALGORITHMS } from "@/lib/algorithms";
import { generateOffsets, simulateAlgorithm, type SimulationParams } from "@/lib/algorithmSimulation";
import type { SimulatedRequest } from "@/lib/types";
import Card from "@/components/ui/Card";
import RequestTimeline from "@/components/RequestTimeline";

interface SimulatorResultsProps {
  params: SimulationParams;
  requestCount: number;
  intervalMs: number;
}

export default function SimulatorResults({ params, requestCount, intervalMs }: SimulatorResultsProps) {
  const offsets = generateOffsets(requestCount, intervalMs);

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
      {ALGORITHMS.map((algorithm) => {
        const decisions = simulateAlgorithm(algorithm.id, offsets, params);

        const results: SimulatedRequest[] = decisions.map((decision, index) => ({
          id: index + 1,
          timestamp: new Date(decision.offsetMs),
          allowed: decision.allowed,
          remaining: decision.remaining,
          retryAfter: decision.retryAfterSeconds,
        }));

        return (
          <Card key={algorithm.id} title={algorithm.name} description={algorithm.summary}>
            <RequestTimeline results={results} formatTime={(r) => `+${offsets[r.id - 1]}ms`} />
          </Card>
        );
      })}
    </div>
  );
}
