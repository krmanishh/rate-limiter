package tokenbucket

import (
	"sync"
	"time"

	"github.com/krmanishh/rate-limiter/internal/limiter"
)

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

type TokenBucketLimiter struct {
	mu         sync.Mutex
	capacity   float64
	refillRate float64
	buckets    map[string]*bucket
}

func New(capacity int, refillRate float64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		capacity:   float64(capacity),
		refillRate: refillRate,
		buckets:    make(map[string]*bucket),
	}
}

func (t *TokenBucketLimiter) Allow(key string) limiter.Result {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	currentBucket, exists := t.buckets[key]

	if !exists {
		currentBucket = &bucket{
			tokens:     t.capacity,
			lastRefill: now,
		}

		t.buckets[key] = currentBucket
	}

	t.refill(currentBucket, now)

	if currentBucket.tokens < 1 {
		retryAfter := time.Duration(
			((1 - currentBucket.tokens) / t.refillRate) * float64(time.Second),
		)

		return limiter.Result{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: limiter.CeilSeconds(retryAfter),
			Limit:      int(t.capacity),
		}
	}

	currentBucket.tokens--

	return limiter.Result{
		Allowed:   true,
		Remaining: int(currentBucket.tokens),
		Limit:     int(t.capacity),
	}
}

func (t *TokenBucketLimiter) refill(
	currentBucket *bucket,
	now time.Time,
) {
	elapsed := now.Sub(currentBucket.lastRefill)

	tokensToAdd := elapsed.Seconds() * t.refillRate

	currentBucket.tokens += tokensToAdd

	if currentBucket.tokens > t.capacity {
		currentBucket.tokens = t.capacity
	}

	currentBucket.lastRefill = now
}
