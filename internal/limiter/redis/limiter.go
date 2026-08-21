package redislimiter

import (
	"context"

	"github.com/krmanishh/rate-limiter/internal/limiter"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string) (limiter.Result, error)
}
