package leakybucket

import (
	"sync"
	"time"

	"github.com/krmanishh/rate-limiter/internal/limiter"
)

type bucket struct {
	queueSize int
	lastLeak  time.Time
}

type LeakyBucketLimiter struct {
	mu       sync.Mutex
	capacity int
	leakRate float64
	buckets  map[string]*bucket
}

func New(capacity int, leakRate float64) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{
		capacity: capacity,
		leakRate: leakRate,
		buckets:  make(map[string]*bucket),
	}
}

func (l *LeakyBucketLimiter) Allow(key string) limiter.Result {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	currentBucket, exists := l.buckets[key]

	if !exists {
		currentBucket = &bucket{
			lastLeak: now,
		}

		l.buckets[key] = currentBucket
	}

	l.leak(currentBucket, now)

	if currentBucket.queueSize >= l.capacity {
		return limiter.Result{
			Allowed:   false,
			Remaining: 0,
			Limit:     l.capacity,
		}
	}

	currentBucket.queueSize++

	return limiter.Result{
		Allowed:   true,
		Remaining: l.capacity - currentBucket.queueSize,
		Limit:     l.capacity,
	}
}

func (l *LeakyBucketLimiter) leak(
	currentBucket *bucket,
	now time.Time,
) {
	elapsed := now.Sub(currentBucket.lastLeak)

	requestsLeaked := int(elapsed.Seconds() * l.leakRate)

	if requestsLeaked <= 0 {
		return
	}

	currentBucket.queueSize -= requestsLeaked

	if currentBucket.queueSize < 0 {
		currentBucket.queueSize = 0
	}

	currentBucket.lastLeak = now
}
