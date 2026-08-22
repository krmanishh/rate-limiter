package fixedwindow

import (
	"sync"
	"time"

	"github.com/krmanishh/rate-limiter/internal/limiter"
)

type window struct {
	count       int
	windowStart time.Time
}

type FixedWindowLimiter struct {
	mu         sync.Mutex
	limit      int
	windowSize time.Duration
	windows    map[string]*window
}

func New(limit int, windowSize time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		limit:      limit,
		windowSize: windowSize,
		windows:    make(map[string]*window),
	}
}

func (f *FixedWindowLimiter) Allow(key string) limiter.Result {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()

	currentWindow, exists := f.windows[key]

	if !exists {
		f.windows[key] = &window{
			count:       1,
			windowStart: now,
		}

		return limiter.Result{
			Allowed:   true,
			Remaining: f.limit - 1,
			Limit:     f.limit,
		}
	}

	if now.Sub(currentWindow.windowStart) >= f.windowSize {
		currentWindow.count = 1
		currentWindow.windowStart = now

		return limiter.Result{
			Allowed:   true,
			Remaining: f.limit - 1,
			Limit:     f.limit,
		}
	}

	if currentWindow.count >= f.limit {
		retryAfter := f.windowSize - now.Sub(currentWindow.windowStart)

		return limiter.Result{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: limiter.CeilSeconds(retryAfter),
			Limit:      f.limit,
		}
	}

	currentWindow.count++

	return limiter.Result{
		Allowed:   true,
		Remaining: f.limit - currentWindow.count,
		Limit:     f.limit,
	}
}
