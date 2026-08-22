package slidingcounter

import (
	"context"
	"fmt"
	"time"

	"github.com/krmanishh/rate-limiter/internal/limiter"
	redislimiter "github.com/krmanishh/rate-limiter/internal/limiter/redis"
	"github.com/krmanishh/rate-limiter/internal/store"
)

const allowScript = `
local key = KEYS[1]
local window_ms = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])

local time = redis.call("TIME")
local now_ms = tonumber(time[1]) * 1000 + math.floor(tonumber(time[2]) / 1000)

local data = redis.call("HMGET", key, "current", "previous", "start")
local current = tonumber(data[1])
local previous = tonumber(data[2])
local window_start = tonumber(data[3])

if current == nil then
	current = 0
	previous = 0
	window_start = now_ms
end

local elapsed = now_ms - window_start

if elapsed >= window_ms and elapsed < 2 * window_ms then
	previous = current
	current = 0
	window_start = window_start + window_ms
	elapsed = now_ms - window_start
elseif elapsed >= 2 * window_ms then
	previous = 0
	current = 0
	window_start = now_ms
	elapsed = 0
end

local previous_weight = (window_ms - elapsed) / window_ms
local estimated = math.floor((previous * previous_weight) + current)

if estimated >= limit then
	redis.call("HSET", key, "current", current, "previous", previous, "start", window_start)
	redis.call("PEXPIRE", key, window_ms * 2)

	return {0, 0}
end

current = current + 1
estimated = estimated + 1

redis.call("HSET", key, "current", current, "previous", previous, "start", window_start)
redis.call("PEXPIRE", key, window_ms * 2)

local remaining = limit - estimated

if remaining < 0 then
	remaining = 0
end

return {1, remaining}
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
		"rate-limit:slidingcounter:%s",
		key,
	)

	result, err := r.store.Eval(
		ctx,
		allowScript,
		[]string{redisKey},
		int64(r.window.Milliseconds()),
		r.limit,
	)

	if err != nil {
		return limiter.Result{}, err
	}

	values, err := redislimiter.ParseInts(result, 2)

	if err != nil {
		return limiter.Result{}, err
	}

	allowed, remaining := values[0], values[1]

	return limiter.Result{
		Allowed:   allowed == 1,
		Remaining: int(remaining),
		Limit:     int(r.limit),
	}, nil
}
