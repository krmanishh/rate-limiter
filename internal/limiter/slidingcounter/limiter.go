package slidingcounter

import (
	"sync"
	"time"

	"github.com/krmanishh/rate-limiter/internal/limiter"
)

type counter struct {
	currentCount  int
	previousCount int
	windowStart   time.Time
}

type SlidingCounterLimiter struct {
	mu         sync.Mutex
	limit      int
	windowSize time.Duration
	counters   map[string]*counter
}

func New(limit int, windowSize time.Duration) *SlidingCounterLimiter {
	return &SlidingCounterLimiter{
		limit:      limit,
		windowSize: windowSize,
		counters:   make(map[string]*counter),
	}
}

func (s *SlidingCounterLimiter) Allow(key string) limiter.Result {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	current, exists := s.counters[key]

	if !exists {
		s.counters[key] = &counter{
			currentCount: 1,
			windowStart:  now,
		}

		return limiter.Result{
			Allowed:   true,
			Remaining: s.limit - 1,
		}
	}

	elapsed := now.Sub(current.windowStart)

	// The current fixed window has not expired.
	if elapsed < s.windowSize {
		estimatedCount := s.estimatedCount(current, elapsed)

		if estimatedCount >= s.limit {
			return limiter.Result{
				Allowed:   false,
				Remaining: 0,
			}
		}

		current.currentCount++

		estimatedCount++

		return limiter.Result{
			Allowed:   true,
			Remaining: s.remaining(estimatedCount),
		}
	}

	// The current window has expired.
	if elapsed < 2*s.windowSize {
		current.previousCount = current.currentCount
		current.currentCount = 0
		current.windowStart = current.windowStart.Add(s.windowSize)

		elapsed = now.Sub(current.windowStart)
	} else {
		// More than one complete window has passed.
		current.previousCount = 0
		current.currentCount = 0
		current.windowStart = now

		elapsed = 0
	}

	estimatedCount := s.estimatedCount(current, elapsed)

	if estimatedCount >= s.limit {
		return limiter.Result{
			Allowed:   false,
			Remaining: 0,
		}
	}

	current.currentCount++
	estimatedCount++

	return limiter.Result{
		Allowed:   true,
		Remaining: s.remaining(estimatedCount),
	}
}

func (s *SlidingCounterLimiter) estimatedCount(
	current *counter,
	elapsed time.Duration,
) int {
	previousWeight := float64(s.windowSize-elapsed) /
		float64(s.windowSize)

	estimated := float64(current.previousCount)*previousWeight +
		float64(current.currentCount)

	return int(estimated)
}

func (s *SlidingCounterLimiter) remaining(count int) int {
	remaining := s.limit - count

	if remaining < 0 {
		return 0
	}

	return remaining
}
