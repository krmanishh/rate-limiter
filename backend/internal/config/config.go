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

type Storage string

const (
	Memory Storage = "memory"
	Redis  Storage = "redis"
)

// FailMode decides what happens when the Redis backend errors (e.g.
// it's unreachable) and a decision can't be reached:
//
//   - FailClosed (default): reject the request. Safer for protecting
//     an expensive or fragile backend — an outage degrades to "nothing
//     gets through" rather than "no limit is enforced."
//   - FailOpen: allow the request. Prioritizes availability — an
//     outage degrades to "unlimited" rather than "fully blocked."
//
// Only relevant when Storage is Redis; memory limiters never error.
type FailMode string

const (
	FailClosed FailMode = "closed"
	FailOpen   FailMode = "open"
)

type RateLimitConfig struct {
	Algorithm Algorithm
	Storage   Storage
	FailMode  FailMode

	RedisAddress string
	// RedisPassword authenticates via Redis AUTH. Empty means no
	// AUTH — fine for a local/dev Redis, but a real deployment should
	// set this (see README security notes on Redis network exposure).
	RedisPassword string

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
