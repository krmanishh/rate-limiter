package leakybucket

import (
	"testing"
	"time"
)

func TestLeakyBucket_AllowsRequestsWithinCapacity(t *testing.T) {
	limiter := New(3, 1)

	for i := 0; i < 3; i++ {
		result := limiter.Allow("user-1")

		if !result.Allowed {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}
}

func TestLeakyBucket_RejectsWhenBucketIsFull(t *testing.T) {
	limiter := New(3, 1)

	for i := 0; i < 3; i++ {
		limiter.Allow("user-1")
	}

	result := limiter.Allow("user-1")

	if result.Allowed {
		t.Fatal("request should have been rejected when bucket is full")
	}

	if result.Remaining != 0 {
		t.Fatalf("expected remaining to be 0, got %d", result.Remaining)
	}
}

func TestLeakyBucket_LeaksOverTime(t *testing.T) {
	limiter := New(2, 20)

	// Fill the bucket.
	if !limiter.Allow("user-1").Allowed {
		t.Fatal("first request should have been allowed")
	}

	if !limiter.Allow("user-1").Allowed {
		t.Fatal("second request should have been allowed")
	}

	// 20 requests/sec means that after 100ms,
	// approximately 2 requests should have leaked.
	time.Sleep(100 * time.Millisecond)

	if !limiter.Allow("user-1").Allowed {
		t.Fatal("request should have been allowed after leakage")
	}
}

func TestLeakyBucket_SeparatesKeys(t *testing.T) {
	limiter := New(1, 1)

	result1 := limiter.Allow("user-1")
	result2 := limiter.Allow("user-2")

	if !result1.Allowed {
		t.Fatal("user-1 request should have been allowed")
	}

	if !result2.Allowed {
		t.Fatal("user-2 request should have been allowed")
	}
}

func TestLeakyBucket_ConcurrentRequests(t *testing.T) {
	limiter := New(100, 1)

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
