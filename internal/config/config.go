package config

import "time"

type Algorithm string

const (
	FixedWindow    Algorithm = "fixed_window"
	SlidingLog     Algorithm = "sliding_log"
	SlidingCounter Algorithm = "sliding_counter"
	TokenBucket    Algorithm = "token_bucket"
	LeakyBucket    Algorithm = "leaky_bucket"
)

type RateLimitConfig struct {
	Algorithm Algorithm

	// Used by Fixed Window, Sliding Log,
	// and Sliding Counter.
	Limit      int
	WindowSize time.Duration

	// Used by Token Bucket.
	Capacity   int
	RefillRate float64

	// Used by Leaky Bucket.
	LeakRate float64
}
