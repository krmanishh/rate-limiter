package leakybucket

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
local leak_rate = tonumber(ARGV[2])

local time = redis.call("TIME")
local now_ms = tonumber(time[1]) * 1000 + math.floor(tonumber(time[2]) / 1000)

local data = redis.call("HMGET", key, "queue", "last_leak")
local queue = tonumber(data[1])
local last_leak = tonumber(data[2])

if queue == nil then
	queue = 0
	last_leak = now_ms
end

local elapsed_seconds = (now_ms - last_leak) / 1000
local leaked = math.floor(elapsed_seconds * leak_rate)

if leaked > 0 then
	queue = queue - leaked

	if queue < 0 then
		queue = 0
	end

	last_leak = now_ms
end

local ttl_ms = math.ceil((capacity / leak_rate) * 1000) + 1000

if queue >= capacity then
	redis.call("HSET", key, "queue", queue, "last_leak", last_leak)
	redis.call("PEXPIRE", key, ttl_ms)

	return {0, 0}
end

queue = queue + 1

redis.call("HSET", key, "queue", queue, "last_leak", last_leak)
redis.call("PEXPIRE", key, ttl_ms)

return {1, capacity - queue}
`

type RedisLimiter struct {
	store    store.Store
	capacity int64
	leakRate float64
}

func NewRedis(
	store store.Store,
	capacity int64,
	leakRate float64,
) *RedisLimiter {
	return &RedisLimiter{
		store:    store,
		capacity: capacity,
		leakRate: leakRate,
	}
}

func (r *RedisLimiter) Allow(
	ctx context.Context,
	key string,
) (limiter.Result, error) {
	redisKey := fmt.Sprintf(
		"rate-limit:leakybucket:%s",
		key,
	)

	result, err := r.store.Eval(
		ctx,
		allowScript,
		[]string{redisKey},
		r.capacity,
		r.leakRate,
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
		Limit:     int(r.capacity),
	}, nil
}
