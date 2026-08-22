package adapter

import (
	"context"
	"log/slog"

	"github.com/krmanishh/rate-limiter/internal/limiter"
	redislimiter "github.com/krmanishh/rate-limiter/internal/limiter/redis"
	"github.com/krmanishh/rate-limiter/internal/metrics"
)

// errorRetryAfterSeconds is a fixed Retry-After hint given on a
// fail-closed rejection caused by a Redis error, since the real
// window/refill timing can't be known without a working Redis.
const errorRetryAfterSeconds = 5

// RedisAdapter adapts a context-aware, error-returning Redis limiter to
// the plain limiter.RateLimiter interface. This is the one place that
// can tell a genuine Redis failure apart from a legitimate rejection,
// so it's also the one place outside the metrics package that records
// rate_limit_errors_total — the algorithm implementations themselves
// never reference Prometheus. It's also where the configured
// fail-open/fail-closed policy is applied.
type RedisAdapter struct {
	limiter   redislimiter.RateLimiter
	algorithm string
	limit     int
	failOpen  bool
}

// NewRedis wraps limiter. limit is the algorithm's configured
// limit/capacity, used only to populate Result.Limit and (in fail-open
// mode) Result.Remaining when Redis errors — the underlying limiter
// isn't consulted for that value in the error path since it couldn't
// answer. failOpen selects the behavior on a Redis error: true allows
// the request (prioritizing availability), false rejects it
// (prioritizing whatever the rate limit is protecting).
func NewRedis(
	limiter redislimiter.RateLimiter,
	algorithm string,
	limit int,
	failOpen bool,
) *RedisAdapter {
	return &RedisAdapter{
		limiter:   limiter,
		algorithm: algorithm,
		limit:     limit,
		failOpen:  failOpen,
	}
}

func (r *RedisAdapter) Allow(key string) limiter.Result {
	result, err := r.limiter.Allow(
		context.Background(),
		key,
	)

	if err != nil {
		metrics.RateLimitErrorsTotal.WithLabelValues(r.algorithm, "redis").Inc()

		slog.Warn(
			"redis rate limiter error, applying fail mode",
			"algorithm", r.algorithm,
			"fail_open", r.failOpen,
			"error", err,
		)

		if r.failOpen {
			return limiter.Result{
				Allowed:   true,
				Remaining: r.limit,
				Limit:     r.limit,
			}
		}

		return limiter.Result{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: errorRetryAfterSeconds,
			Limit:      r.limit,
		}
	}

	return result
}
