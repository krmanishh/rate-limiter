package slidinglog

import (
	"testing"
	"time"
)

func TestSlidingLog_AllowsRequestsWithinLimit(t *testing.T) {
	limiter := New(3, time.Second)

	for i := 0; i < 3; i++ {
		result := limiter.Allow("user-1")

		if !result.Allowed {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}
}

func TestSlidingLog_RejectsAfterLimit(t *testing.T) {
	limiter := New(3, time.Second)

	for i := 0; i < 3; i++ {
		limiter.Allow("user-1")
	}

	result := limiter.Allow("user-1")

	if result.Allowed {
		t.Fatal("4th request should have been rejected")
	}

	if result.Remaining != 0 {
		t.Fatalf("expected remaining to be 0, got %d", result.Remaining)
	}
}

func TestSlidingLog_SeparatesKeys(t *testing.T) {
	limiter := New(2, time.Second)

	result1 := limiter.Allow("user-1")
	result2 := limiter.Allow("user-2")

	if !result1.Allowed {
		t.Fatal("user-1 request should have been allowed")
	}

	if !result2.Allowed {
		t.Fatal("user-2 request should have been allowed")
	}
}

func TestSlidingLog_ExpiredTimestampsAreRemoved(t *testing.T) {
	limiter := New(1, 50*time.Millisecond)

	result := limiter.Allow("user-1")

	if !result.Allowed {
		t.Fatal("first request should have been allowed")
	}

	result = limiter.Allow("user-1")

	if result.Allowed {
		t.Fatal("second request should have been rejected")
	}

	time.Sleep(60 * time.Millisecond)

	result = limiter.Allow("user-1")

	if !result.Allowed {
		t.Fatal("request should have been allowed after timestamp expired")
	}
}

func TestSlidingLog_WindowSlides(t *testing.T) {
	limiter := New(2, 100*time.Millisecond)

	result := limiter.Allow("user-1")

	if !result.Allowed {
		t.Fatal("first request should have been allowed")
	}

	time.Sleep(50 * time.Millisecond)

	result = limiter.Allow("user-1")

	if !result.Allowed {
		t.Fatal("second request should have been allowed")
	}

	result = limiter.Allow("user-1")

	if result.Allowed {
		t.Fatal("third request should have been rejected")
	}

	time.Sleep(60 * time.Millisecond)

	// The first request is now older than 100ms.
	// The second request is still inside the window.
	result = limiter.Allow("user-1")

	if !result.Allowed {
		t.Fatal("request should have been allowed after the window slid")
	}
}

func TestSlidingLog_ConcurrentRequests(t *testing.T) {
	limiter := New(100, time.Second)

	const totalRequests = 100

	results := make(chan bool, totalRequests)

	for i := 0; i < totalRequests; i++ {
		go func() {
			result := limiter.Allow("user-1")
			results <- result.Allowed
		}()
	}

	allowedCount := 0

	for i := 0; i < totalRequests; i++ {
		if <-results {
			allowedCount++
		}
	}

	if allowedCount != totalRequests {
		t.Fatalf(
			"expected %d allowed requests, got %d",
			totalRequests,
			allowedCount,
		)
	}
}
