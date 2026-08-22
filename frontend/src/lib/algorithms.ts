import type { Algorithm, ConfigResponse } from "./types";

export type ConfigFieldKey =
  | "limit"
  | "window_seconds"
  | "capacity"
  | "refill_rate"
  | "leak_rate";

export interface ConfigFieldMeta {
  key: ConfigFieldKey;
  label: string;
  unit: string;
}

export interface AlgorithmMeta {
  id: Algorithm;
  name: string;
  summary: string;
  description: string;
  pros: string[];
  cons: string[];
  /** Which fields of ConfigResponse this algorithm actually uses. */
  fields: ConfigFieldMeta[];
}

export const ALGORITHMS: AlgorithmMeta[] = [
  {
    id: "fixed_window",
    name: "Fixed Window",
    summary: "Count requests in a fixed-size time bucket.",
    description:
      "Counts requests in a fixed-size time bucket (e.g. \"this minute\"). When the bucket rolls over, the count resets to zero. Simple to reason about and cheap to implement — a single counter per key.",
    pros: [
      "Simple to implement and reason about",
      "O(1) memory per key",
      "Cheap: a single counter and an expiry",
    ],
    cons: [
      "Allows up to 2x the limit at window boundaries (a client can spend its full quota at the end of one window and again at the start of the next)",
      "Not a true sliding window",
    ],
    fields: [
      { key: "limit", label: "Limit", unit: "requests" },
      { key: "window_seconds", label: "Window", unit: "seconds" },
    ],
  },
  {
    id: "sliding_log",
    name: "Sliding Log",
    summary: "Keep a timestamp per request; count how many fall in the trailing window.",
    description:
      "Keeps a timestamp for every request and counts how many fall inside the trailing window (\"the last 60 seconds\", continuously, not a fixed bucket). Exact — no boundary bursting — at the cost of memory that scales with request volume.",
    pros: [
      "Exact: no boundary-burst problem",
      "Retry-After is precise (time until the oldest logged request ages out)",
    ],
    cons: [
      "Memory scales with request volume, not O(1) per key",
      "More expensive per check (pruning old entries, counting)",
    ],
    fields: [
      { key: "limit", label: "Limit", unit: "requests" },
      { key: "window_seconds", label: "Window", unit: "seconds" },
    ],
  },
  {
    id: "sliding_counter",
    name: "Sliding Counter",
    summary: "Approximate a sliding window with a weighted average of two fixed windows.",
    description:
      "Approximates a sliding window cheaply: it keeps counts for the current and previous fixed windows, and estimates the \"true\" sliding count as a weighted average based on how far into the current window we are. O(1) memory, close to sliding-log accuracy.",
    pros: [
      "O(1) memory per key, unlike sliding log",
      "Much less boundary bursting than fixed window",
    ],
    cons: [
      "An approximation, not exact — the weighted estimate can be slightly off",
      "Retry-After is an estimate, not a guarantee, for the same reason",
    ],
    fields: [
      { key: "limit", label: "Limit", unit: "requests" },
      { key: "window_seconds", label: "Window", unit: "seconds" },
    ],
  },
  {
    id: "token_bucket",
    name: "Token Bucket",
    summary: "A bucket refills at a fixed rate; each request consumes a token.",
    description:
      "A bucket holds up to `capacity` tokens and refills at `refill_rate` tokens/second. Each request consumes one token; if none are available, the request is rejected. Allows controlled bursts up to the bucket's capacity while enforcing a steady average rate over time.",
    pros: [
      "Allows legitimate bursts up to capacity",
      "Smooth, well-understood behavior; widely used in practice",
    ],
    cons: [
      "Two parameters to tune (capacity and refill rate) instead of one",
      "A burst can still momentarily exceed what a fixed-window limit of the same average rate would allow",
    ],
    fields: [
      { key: "capacity", label: "Capacity", unit: "tokens" },
      { key: "refill_rate", label: "Refill rate", unit: "tokens/sec" },
    ],
  },
  {
    id: "leaky_bucket",
    name: "Leaky Bucket",
    summary: "A queue drains at a fixed rate; each request adds to the queue.",
    description:
      "A queue (bucket) fills as requests arrive and drains at a fixed `leak_rate` requests/second. If the queue is full, new requests are rejected. Unlike token bucket, it smooths bursts into a steady outflow rather than allowing them through immediately.",
    pros: [
      "Produces a smooth, steady outbound rate — good for protecting a downstream system that can't handle bursts",
      "Predictable worst-case behavior",
    ],
    cons: [
      "Doesn't allow bursts at all, even brief legitimate ones",
      "Two parameters to tune (capacity and leak rate)",
    ],
    fields: [
      { key: "capacity", label: "Capacity", unit: "slots" },
      { key: "leak_rate", label: "Leak rate", unit: "requests/sec" },
    ],
  },
];

export function getAlgorithm(id: Algorithm): AlgorithmMeta {
  const found = ALGORITHMS.find((algorithm) => algorithm.id === id);

  if (!found) {
    throw new Error(`unknown algorithm: ${id}`);
  }

  return found;
}

export function algorithmLabel(id: Algorithm): string {
  return getAlgorithm(id).name;
}

/** Reads a ConfigFieldMeta's value out of a live ConfigResponse. */
export function fieldValue(field: ConfigFieldKey, config: ConfigResponse): number {
  return config[field];
}
