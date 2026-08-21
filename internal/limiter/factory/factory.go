package factory

import (
	"fmt"

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

func Create(cfg config.RateLimitConfig) (limiter.RateLimiter, error) {
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
		return nil, fmt.Errorf(
			"unsupported rate limiter algorithm: %s",
			cfg.Algorithm,
		)
	}
}

func fixedWindow(
	cfg config.RateLimitConfig,
) (limiter.RateLimiter, error) {
	if cfg.Limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than 0")
	}

	if cfg.WindowSize <= 0 {
		return nil, fmt.Errorf("window size must be greater than 0")
	}

	if cfg.Storage == config.Redis {
		redisStore, err := newRedisStore(cfg)

		if err != nil {
			return nil, err
		}

		redisLimiter := fixedwindow.NewRedis(
			redisStore,
			int64(cfg.Limit),
			cfg.WindowSize,
		)

		return adapter.NewRedis(redisLimiter), nil
	}

	return fixedwindow.New(
		cfg.Limit,
		cfg.WindowSize,
	), nil
}

func slidingLog(
	cfg config.RateLimitConfig,
) (limiter.RateLimiter, error) {
	if cfg.Limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than 0")
	}

	if cfg.WindowSize <= 0 {
		return nil, fmt.Errorf("window size must be greater than 0")
	}

	if cfg.Storage == config.Redis {
		redisStore, err := newRedisStore(cfg)

		if err != nil {
			return nil, err
		}

		redisLimiter := slidinglog.NewRedis(
			redisStore,
			int64(cfg.Limit),
			cfg.WindowSize,
		)

		return adapter.NewRedis(redisLimiter), nil
	}

	return slidinglog.New(
		cfg.Limit,
		cfg.WindowSize,
	), nil
}

func slidingCounter(
	cfg config.RateLimitConfig,
) (limiter.RateLimiter, error) {
	if cfg.Limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than 0")
	}

	if cfg.WindowSize <= 0 {
		return nil, fmt.Errorf("window size must be greater than 0")
	}

	if cfg.Storage == config.Redis {
		redisStore, err := newRedisStore(cfg)

		if err != nil {
			return nil, err
		}

		redisLimiter := slidingcounter.NewRedis(
			redisStore,
			int64(cfg.Limit),
			cfg.WindowSize,
		)

		return adapter.NewRedis(redisLimiter), nil
	}

	return slidingcounter.New(
		cfg.Limit,
		cfg.WindowSize,
	), nil
}

func tokenBucket(
	cfg config.RateLimitConfig,
) (limiter.RateLimiter, error) {
	if cfg.Capacity <= 0 {
		return nil, fmt.Errorf("capacity must be greater than 0")
	}

	if cfg.RefillRate <= 0 {
		return nil, fmt.Errorf("refill rate must be greater than 0")
	}

	if cfg.Storage == config.Redis {
		redisStore, err := newRedisStore(cfg)

		if err != nil {
			return nil, err
		}

		redisLimiter := tokenbucket.NewRedis(
			redisStore,
			int64(cfg.Capacity),
			cfg.RefillRate,
		)

		return adapter.NewRedis(redisLimiter), nil
	}

	return tokenbucket.New(
		cfg.Capacity,
		cfg.RefillRate,
	), nil
}

func leakyBucket(
	cfg config.RateLimitConfig,
) (limiter.RateLimiter, error) {
	if cfg.Capacity <= 0 {
		return nil, fmt.Errorf("capacity must be greater than 0")
	}

	if cfg.LeakRate <= 0 {
		return nil, fmt.Errorf("leak rate must be greater than 0")
	}

	if cfg.Storage == config.Redis {
		redisStore, err := newRedisStore(cfg)

		if err != nil {
			return nil, err
		}

		redisLimiter := leakybucket.NewRedis(
			redisStore,
			int64(cfg.Capacity),
			cfg.LeakRate,
		)

		return adapter.NewRedis(redisLimiter), nil
	}

	return leakybucket.New(
		cfg.Capacity,
		cfg.LeakRate,
	), nil
}

func newRedisStore(
	cfg config.RateLimitConfig,
) (*redisstore.Client, error) {
	if cfg.RedisAddress == "" {
		return nil, fmt.Errorf("redis address must be set when storage is redis")
	}

	return redisstore.New(cfg.RedisAddress), nil
}
