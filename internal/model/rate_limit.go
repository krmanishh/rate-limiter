package model

type RateLimitRequest struct {
	Key string `json:"key"`
}

type RateLimitResponse struct {
	Allowed    bool `json:"allowed"`
	Remaining  int  `json:"remaining"`
	RetryAfter int  `json:"retry_after"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// ConfigResponse exposes the server's active rate limit configuration
// (loaded from environment variables at startup) so a client can render
// it without needing separate access to the server's environment.
// Redis connection details are deliberately omitted — a client has no
// legitimate need for them, and RedisPassword must never be exposed.
type ConfigResponse struct {
	Algorithm     string  `json:"algorithm"`
	Storage       string  `json:"storage"`
	FailMode      string  `json:"fail_mode"`
	Limit         int     `json:"limit"`
	WindowSeconds int     `json:"window_seconds"`
	Capacity      int     `json:"capacity"`
	RefillRate    float64 `json:"refill_rate"`
	LeakRate      float64 `json:"leak_rate"`
}
