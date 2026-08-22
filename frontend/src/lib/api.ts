import type {
  ConfigResponse,
  ErrorResponse,
  HealthResponse,
  RateLimitCheckResponse,
} from "./types";

export const API_BASE_URL = (
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080"
).replace(/\/$/, "");

export class ApiError extends Error {
  readonly status?: number;

  constructor(message: string, status?: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

async function parseErrorBody(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as ErrorResponse;

    if (body && typeof body.error === "string" && body.error.length > 0) {
      return body.error;
    }
  } catch {
    // Body wasn't JSON (or was empty) — fall through to a generic message.
  }

  return `Request failed with status ${response.status}`;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let response: Response;

  try {
    response = await fetch(`${API_BASE_URL}${path}`, init);
  } catch {
    throw new ApiError(
      `Could not reach the API at ${API_BASE_URL}. Is the Go server running, and is CORS_ALLOWED_ORIGIN set to this app's origin?`,
    );
  }

  if (!response.ok) {
    throw new ApiError(await parseErrorBody(response), response.status);
  }

  return (await response.json()) as T;
}

export function fetchHealth(signal?: AbortSignal): Promise<HealthResponse> {
  return request<HealthResponse>("/health", { signal });
}

export function fetchConfig(signal?: AbortSignal): Promise<ConfigResponse> {
  return request<ConfigResponse>("/api/v1/config", { signal });
}

export function checkRateLimit(
  key: string,
  signal?: AbortSignal,
): Promise<RateLimitCheckResponse> {
  return request<RateLimitCheckResponse>("/api/v1/ratelimit/check", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ key }),
    signal,
  });
}
