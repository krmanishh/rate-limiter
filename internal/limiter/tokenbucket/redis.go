package tokenbucket

import (
	"context"
	"fmt"

	"github.com/krmanishh/rate-limiter/internal/limiter"
	redislimiter "github.com/krmanishh/rate-limiter/internal/limiter/redis"
	"github.com/krmanishh/rate-limiter/internal/store"
)

const allowScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])

local time = redis.call("TIME")
local now_ms = tonumber(time[1]) * 1000 + math.floor(tonumber(time[2]) / 1000)

local data = redis.call("HMGET", key, "tokens", "last_refill")
local tokens = tonumber(data[1])
local last_refill = tonumber(data[2])

if tokens == nil then
	tokens = capacity
	last_refill = now_ms
end

local elapsed_seconds = (now_ms - last_refill) / 1000
tokens = tokens + (elapsed_seconds * refill_rate)

if tokens > capacity then
	tokens = capacity
end

local ttl_ms = math.ceil((capacity / refill_rate) * 1000) + 1000

if tokens < 1 then
	local deficit = 1 - tokens
	local retry_after_ms = math.ceil((deficit / refill_rate) * 1000)

	redis.call("HSET", key, "tokens", tokens, "last_refill", now_ms)
	redis.call("PEXPIRE", key, ttl_ms)

	return {0, 0, retry_after_ms}
end

tokens = tokens - 1

redis.call("HSET", key, "tokens", tokens, "last_refill", now_ms)
redis.call("PEXPIRE", key, ttl_ms)

return {1, math.floor(tokens), 0}
`

type RedisLimiter struct {
	store      store.Store
	capacity   int64
	refillRate float64
}

func NewRedis(
	store store.Store,
	capacity int64,
	refillRate float64,
) *RedisLimiter {
	return &RedisLimiter{
		store:      store,
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (r *RedisLimiter) Allow(
	ctx context.Context,
	key string,
) (limiter.Result, error) {
	redisKey := fmt.Sprintf(
		"rate-limit:tokenbucket:%s",
		key,
	)

	result, err := r.store.Eval(
		ctx,
		allowScript,
		[]string{redisKey},
		r.capacity,
		r.refillRate,
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
	}, nil
}
