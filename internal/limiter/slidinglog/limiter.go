package slidinglog

import (
	"sync"
	"time"

	"github.com/krmanishh/rate-limiter/internal/limiter"
)

type SlidingLogLimiter struct {
	mu         sync.Mutex
	limit      int
	windowSize time.Duration
	logs       map[string][]time.Time
}

func New(limit int, windowSize time.Duration) *SlidingLogLimiter {
	return &SlidingLogLimiter{
		limit:      limit,
		windowSize: windowSize,
		logs:       make(map[string][]time.Time),
	}
}

func (s *SlidingLogLimiter) Allow(key string) limiter.Result {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-s.windowSize)

	timestamps := s.logs[key]

	// Remove timestamps outside the sliding window.
	firstValidIndex := 0

	for firstValidIndex < len(timestamps) &&
		timestamps[firstValidIndex].Before(windowStart) {
		firstValidIndex++
	}

	timestamps = timestamps[firstValidIndex:]

	// If the request limit has already been reached,
	// reject the request.
	if len(timestamps) >= s.limit {
		s.logs[key] = timestamps

		retryAfter := timestamps[0].Add(s.windowSize).Sub(now)

		return limiter.Result{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: limiter.CeilSeconds(retryAfter),
			Limit:      s.limit,
		}
	}

	// Add the current request timestamp.
	timestamps = append(timestamps, now)

	s.logs[key] = timestamps

	return limiter.Result{
		Allowed:   true,
		Remaining: s.limit - len(timestamps),
		Limit:     s.limit,
	}
}
