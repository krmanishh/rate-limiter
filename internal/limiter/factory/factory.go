package factory

import (
	"fmt"

	"github.com/krmanishh/rate-limiter/internal/config"
	"github.com/krmanishh/rate-limiter/internal/limiter"
	"github.com/krmanishh/rate-limiter/internal/limiter/fixedwindow"
	"github.com/krmanishh/rate-limiter/internal/limiter/leakybucket"
	"github.com/krmanishh/rate-limiter/internal/limiter/slidingcounter"
	"github.com/krmanishh/rate-limiter/internal/limiter/slidinglog"
	"github.com/krmanishh/rate-limiter/internal/limiter/tokenbucket"
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

	return leakybucket.New(
		cfg.Capacity,
		cfg.LeakRate,
	), nil
}
