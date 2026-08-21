package fixedwindow

import (
	"context"
	"fmt"
	"time"

	"github.com/krmanishh/rate-limiter/internal/limiter"
	"github.com/krmanishh/rate-limiter/internal/store"
)

const incrementScript = `
local current = redis.call("INCR", KEYS[1])

if current == 1 then
	redis.call("EXPIRE", KEYS[1], ARGV[1])
end

return current
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

	current, ok := result.(int64)

	if !ok {
		return limiter.Result{}, fmt.Errorf(
			"unexpected Redis result type: %T",
			result,
		)
	}

	remaining := r.limit - current

	if remaining < 0 {
		remaining = 0
	}

	return limiter.Result{
		Allowed:    current <= r.limit,
		Remaining:  int(remaining),
		RetryAfter: 0,
	}, nil
}
