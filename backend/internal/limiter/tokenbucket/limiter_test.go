package tokenbucket

import (
	"testing"
	"time"
)

func TestTokenBucket_AllowsBurstUpToCapacity(t *testing.T) {
	limiter := New(5, 1)

	for i := 0; i < 5; i++ {
		result := limiter.Allow("user-1")

		if !result.Allowed {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}

	result := limiter.Allow("user-1")

	if result.Allowed {
		t.Fatal("request beyond bucket capacity should have been rejected")
	}
}

func TestTokenBucket_RefillsTokens(t *testing.T) {
	limiter := New(2, 20)

	// Consume both initial tokens.
	if !limiter.Allow("user-1").Allowed {
		t.Fatal("first request should have been allowed")
	}

	if !limiter.Allow("user-1").Allowed {
		t.Fatal("second request should have been allowed")
	}

	// Refill rate = 20 tokens/sec.
	// 100ms should add approximately 2 tokens.
	time.Sleep(100 * time.Millisecond)

	if !limiter.Allow("user-1").Allowed {
		t.Fatal("request should have been allowed after refill")
	}
}

func TestTokenBucket_DoesNotExceedCapacity(t *testing.T) {
	limiter := New(2, 100)

	// Consume both tokens.
	limiter.Allow("user-1")
	limiter.Allow("user-1")

	// Wait long enough for more than two tokens to theoretically arrive.
	time.Sleep(100 * time.Millisecond)

	// Capacity is still only 2, so exactly two requests
	// should be available.
	if !limiter.Allow("user-1").Allowed {
		t.Fatal("first refilled request should have been allowed")
	}

	if !limiter.Allow("user-1").Allowed {
		t.Fatal("second refilled request should have been allowed")
	}

	if limiter.Allow("user-1").Allowed {
		t.Fatal("bucket should not contain more than its capacity")
	}
}

func TestTokenBucket_SeparatesKeys(t *testing.T) {
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

func TestTokenBucket_ConcurrentRequests(t *testing.T) {
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
