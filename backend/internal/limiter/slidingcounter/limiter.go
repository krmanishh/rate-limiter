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
			Limit:     s.limit,
		}
	}

	elapsed := now.Sub(current.windowStart)

	// The current fixed window has not expired.
	if elapsed < s.windowSize {
		estimatedCount := s.estimatedCount(current, elapsed)

		if estimatedCount >= s.limit {
			return limiter.Result{
				Allowed:    false,
				Remaining:  0,
				RetryAfter: limiter.CeilSeconds(s.retryAfter(current, elapsed)),
				Limit:      s.limit,
			}
		}

		current.currentCount++

		estimatedCount++

		return limiter.Result{
			Allowed:   true,
			Remaining: s.remaining(estimatedCount),
			Limit:     s.limit,
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
			Allowed:    false,
			Remaining:  0,
			RetryAfter: limiter.CeilSeconds(s.retryAfter(current, elapsed)),
			Limit:      s.limit,
		}
	}

	current.currentCount++
	estimatedCount++

	return limiter.Result{
		Allowed:   true,
		Remaining: s.remaining(estimatedCount),
		Limit:     s.limit,
	}
}

// retryAfter estimates how long until the weighted estimate decays
// below the limit. previousCount only decays (never grows) as elapsed
// time within the window increases, so if currentCount alone is
// already at or over the limit — or there's no previous window
// contribution to decay — the only thing that can help is the window
// rolling over. Otherwise, solve for the elapsed time at which
// previousCount's linearly-decaying weight drops enough to make room:
//
//	previousCount * (windowSize-elapsed')/windowSize = limit - currentCount
//
// This is an estimate, not a guarantee, because the algorithm itself
// is an approximation of a true sliding window.
func (s *SlidingCounterLimiter) retryAfter(
	current *counter,
	elapsed time.Duration,
) time.Duration {
	remaining := s.limit - current.currentCount

	if remaining <= 0 || current.previousCount <= 0 {
		return s.windowSize - elapsed
	}

	fraction := float64(remaining) / float64(current.previousCount)

	if fraction >= 1 {
		return 0
	}

	targetElapsed := s.windowSize - time.Duration(fraction*float64(s.windowSize))

	retryAfter := targetElapsed - elapsed

	if retryAfter < 0 {
		return 0
	}

	return retryAfter
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
