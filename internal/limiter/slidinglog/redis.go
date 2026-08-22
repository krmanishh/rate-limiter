package slidinglog

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/krmanishh/rate-limiter/internal/limiter"
	redislimiter "github.com/krmanishh/rate-limiter/internal/limiter/redis"
	"github.com/krmanishh/rate-limiter/internal/store"
)

const allowScript = `
local key = KEYS[1]
local window_ms = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local member = ARGV[3]

local time = redis.call("TIME")
local now_ms = tonumber(time[1]) * 1000 + math.floor(tonumber(time[2]) / 1000)
local window_start = now_ms - window_ms

redis.call("ZREMRANGEBYSCORE", key, "-inf", window_start)

local count = redis.call("ZCARD", key)

if count >= limit then
	redis.call("PEXPIRE", key, window_ms)

	local oldest = redis.call("ZRANGE", key, 0, 0, "WITHSCORES")
	local retry_after_ms = 0

	if #oldest == 2 then
		retry_after_ms = (tonumber(oldest[2]) + window_ms) - now_ms
	end

	return {0, 0, retry_after_ms}
end

redis.call("ZADD", key, now_ms, member)
redis.call("PEXPIRE", key, window_ms)

return {1, limit - count - 1, 0}
`

var memberSeq uint64

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
		"rate-limit:slidinglog:%s",
		key,
	)

	member := fmt.Sprintf(
		"%d-%d",
		time.Now().UnixNano(),
		atomic.AddUint64(&memberSeq, 1),
	)

	result, err := r.store.Eval(
		ctx,
		allowScript,
		[]string{redisKey},
		r.window.Milliseconds(),
		r.limit,
		member,
	)

	if err != nil {
		return limiter.Result{}, err
	}

	values, err := redislimiter.ParseInts(result, 3)

	if err != nil {
		return limiter.Result{}, err
	}

	allowed, remaining, retryAfterMs := values[0], values[1], values[2]

	return limiter.Result{
		Allowed:    allowed == 1,
		Remaining:  int(remaining),
		RetryAfter: int(retryAfterMs / 1000),
		Limit:      int(r.limit),
	}, nil
}
