package limiter

import "time"

type Result struct {
	Allowed    bool
	Remaining  int
	RetryAfter int
	Limit      int
}

type RateLimiter interface {
	Allow(key string) Result
}

// CeilSeconds converts d to whole seconds, rounding up. Used for
// Retry-After hints: truncating a sub-second wait down to 0 would tell
// a client to retry before enough time has actually passed.
func CeilSeconds(d time.Duration) int {
	if d <= 0 {
		return 0
	}

	return int((d + time.Second - 1) / time.Second)
}
