package limiter

type Result struct {
	Allowed    bool
	Remaining  int
	RetryAfter int
	Limit      int
}

type RateLimiter interface {
	Allow(key string) Result
}
