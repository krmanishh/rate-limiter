package adapter

import (
	"context"

	"github.com/krmanishh/rate-limiter/internal/limiter"
	redislimiter "github.com/krmanishh/rate-limiter/internal/limiter/redis"
)

type RedisAdapter struct {
	limiter redislimiter.RateLimiter
}

func NewRedis(
	limiter redislimiter.RateLimiter,
) *RedisAdapter {
	return &RedisAdapter{
		limiter: limiter,
	}
}

func (r *RedisAdapter) Allow(key string) limiter.Result {
	result, err := r.limiter.Allow(
		context.Background(),
		key,
	)

	if err != nil {
		return limiter.Result{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: 0,
		}
	}

	return result
}
