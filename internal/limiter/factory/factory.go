package factory

import (
	"fmt"
	"io"

	"github.com/krmanishh/rate-limiter/internal/config"
	"github.com/krmanishh/rate-limiter/internal/limiter"
	"github.com/krmanishh/rate-limiter/internal/limiter/adapter"
	"github.com/krmanishh/rate-limiter/internal/limiter/fixedwindow"
	"github.com/krmanishh/rate-limiter/internal/limiter/leakybucket"
	"github.com/krmanishh/rate-limiter/internal/limiter/slidingcounter"
	"github.com/krmanishh/rate-limiter/internal/limiter/slidinglog"
	"github.com/krmanishh/rate-limiter/internal/limiter/tokenbucket"
	"github.com/krmanishh/rate-limiter/internal/store/redisstore"
)

// noopCloser lets callers always call Close() on whatever Create returns,
// regardless of storage backend, without a type assertion.
type noopCloser struct{}

func (noopCloser) Close() error { return nil }

// Create builds a rate limiter for the given configuration. The returned
// io.Closer owns any resources the limiter opened (e.g. a Redis client)
// and must be closed by the caller once the limiter is no longer needed.
func Create(cfg config.RateLimitConfig) (limiter.RateLimiter, io.Closer, error) {
	switch cfg.Algorithm {
	case config.FixedWindow:
		return fixedWindow(cfg)

	case config.SlidingLog:
		return slidingLog(cfg)

	case config.SlidingCounter:
		return slidingCounter(cfg)

	case config.TokenBucket:
		return tokenBucket(cfg)

	case config.LeakyBucket:
		return leakyBucket(cfg)

	default:
		return nil, nil, fmt.Errorf(
			"unsupported rate limiter algorithm: %s",
			cfg.Algorithm,
		)
	}
}

func fixedWindow(
	cfg config.RateLimitConfig,
) (limiter.RateLimiter, io.Closer, error) {
	if cfg.Limit <= 0 {
		return nil, nil, fmt.Errorf("limit must be greater than 0")
	}

	if cfg.WindowSize <= 0 {
		return nil, nil, fmt.Errorf("window size must be greater than 0")
	}

	if cfg.Storage == config.Redis {
		redisStore, err := newRedisStore(cfg)

		if err != nil {
			return nil, nil, err
		}

		redisLimiter := fixedwindow.NewRedis(
			redisStore,
			int64(cfg.Limit),
			cfg.WindowSize,
		)

		return adapter.NewRedis(redisLimiter, string(cfg.Algorithm), cfg.Limit, failOpen(cfg)), redisStore, nil
	}

	return fixedwindow.New(
		cfg.Limit,
		cfg.WindowSize,
	), noopCloser{}, nil
}

func slidingLog(
	cfg config.RateLimitConfig,
) (limiter.RateLimiter, io.Closer, error) {
	if cfg.Limit <= 0 {
		return nil, nil, fmt.Errorf("limit must be greater than 0")
	}

	if cfg.WindowSize <= 0 {
		return nil, nil, fmt.Errorf("window size must be greater than 0")
	}

	if cfg.Storage == config.Redis {
		redisStore, err := newRedisStore(cfg)

		if err != nil {
			return nil, nil, err
		}

		redisLimiter := slidinglog.NewRedis(
			redisStore,
			int64(cfg.Limit),
			cfg.WindowSize,
		)

		return adapter.NewRedis(redisLimiter, string(cfg.Algorithm), cfg.Limit, failOpen(cfg)), redisStore, nil
	}

	return slidinglog.New(
		cfg.Limit,
		cfg.WindowSize,
	), noopCloser{}, nil
}

func slidingCounter(
	cfg config.RateLimitConfig,
) (limiter.RateLimiter, io.Closer, error) {
	if cfg.Limit <= 0 {
		return nil, nil, fmt.Errorf("limit must be greater than 0")
	}

	if cfg.WindowSize <= 0 {
		return nil, nil, fmt.Errorf("window size must be greater than 0")
	}

	if cfg.Storage == config.Redis {
		redisStore, err := newRedisStore(cfg)

		if err != nil {
			return nil, nil, err
		}

		redisLimiter := slidingcounter.NewRedis(
			redisStore,
			int64(cfg.Limit),
			cfg.WindowSize,
		)

		return adapter.NewRedis(redisLimiter, string(cfg.Algorithm), cfg.Limit, failOpen(cfg)), redisStore, nil
	}

	return slidingcounter.New(
		cfg.Limit,
		cfg.WindowSize,
	), noopCloser{}, nil
}

func tokenBucket(
	cfg config.RateLimitConfig,
) (limiter.RateLimiter, io.Closer, error) {
	if cfg.Capacity <= 0 {
		return nil, nil, fmt.Errorf("capacity must be greater than 0")
	}

	if cfg.RefillRate <= 0 {
		return nil, nil, fmt.Errorf("refill rate must be greater than 0")
	}

	if cfg.Storage == config.Redis {
		redisStore, err := newRedisStore(cfg)

		if err != nil {
			return nil, nil, err
		}

		redisLimiter := tokenbucket.NewRedis(
			redisStore,
			int64(cfg.Capacity),
			cfg.RefillRate,
		)

		return adapter.NewRedis(redisLimiter, string(cfg.Algorithm), cfg.Capacity, failOpen(cfg)), redisStore, nil
	}

	return tokenbucket.New(
		cfg.Capacity,
		cfg.RefillRate,
	), noopCloser{}, nil
}

func leakyBucket(
	cfg config.RateLimitConfig,
) (limiter.RateLimiter, io.Closer, error) {
	if cfg.Capacity <= 0 {
		return nil, nil, fmt.Errorf("capacity must be greater than 0")
	}

	if cfg.LeakRate <= 0 {
		return nil, nil, fmt.Errorf("leak rate must be greater than 0")
	}

	if cfg.Storage == config.Redis {
		redisStore, err := newRedisStore(cfg)

		if err != nil {
			return nil, nil, err
		}

		redisLimiter := leakybucket.NewRedis(
			redisStore,
			int64(cfg.Capacity),
			cfg.LeakRate,
		)

		return adapter.NewRedis(redisLimiter, string(cfg.Algorithm), cfg.Capacity, failOpen(cfg)), redisStore, nil
	}

	return leakybucket.New(
		cfg.Capacity,
		cfg.LeakRate,
	), noopCloser{}, nil
}

func failOpen(cfg config.RateLimitConfig) bool {
	return cfg.FailMode == config.FailOpen
}

func newRedisStore(
	cfg config.RateLimitConfig,
) (*redisstore.Client, error) {
	if cfg.RedisAddress == "" {
		return nil, fmt.Errorf("redis address must be set when storage is redis")
	}

	return redisstore.NewWithPassword(cfg.RedisAddress, cfg.RedisPassword), nil
}
