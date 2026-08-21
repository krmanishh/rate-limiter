package limiter

type Result struct {
	Allowed    bool
	Remaining  int
	RetryAfter int
}

type RateLimiter interface {
	Allow(key string) Result
}