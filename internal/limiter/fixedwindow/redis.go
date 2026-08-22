package fixedwindow

import (
	"context"
	"fmt"
	"time"

	"github.com/krmanishh/rate-limiter/internal/limiter"
	redislimiter "github.com/krmanishh/rate-limiter/internal/limiter/redis"
	"github.com/krmanishh/rate-limiter/internal/store"
)

const incrementScript = `
local current = redis.call("INCR", KEYS[1])

if current == 1 then
	redis.call("EXPIRE", KEYS[1], ARGV[1])
end

local ttl = redis.call("TTL", KEYS[1])

return {current, ttl}
`

type RedisLimiter struct {
	store  store.Store
	limit  int64
	window time.Duration
}

func NewRedis(
	store store.Store,
	limit int64,
	window time.Duration,
) *RedisLimiter {
	return &RedisLimiter{
		store:  store,
		limit:  limit,
		window: window,
	}
}

func (r *RedisLimiter) Allow(
	ctx context.Context,
	key string,
) (limiter.Result, error) {
	redisKey := fmt.Sprintf(
		"rate-limit:fixed:%s",
		key,
	)

	result, err := r.store.Eval(
		ctx,
		incrementScript,
		[]string{redisKey},
		int64(r.window.Seconds()),
	)

	if err != nil {
		return limiter.Result{}, err
	}

	values, err := redislimiter.ParseInts(result, 2)

	if err != nil {
		return limiter.Result{}, err
	}

	current, ttl := values[0], values[1]

	remaining := r.limit - current

	if remaining < 0 {
		remaining = 0
	}

	retryAfter := 0

	if current > r.limit && ttl > 0 {
		retryAfter = int(ttl)
	}

	return limiter.Result{
		Allowed:    current <= r.limit,
		Remaining:  int(remaining),
		RetryAfter: retryAfter,
		Limit:      int(r.limit),
	}, nil
}
