package fixedwindow

import (
	"testing"
	"time"
)

func TestFixedWindow_AllowsRequestsWithinLimit(t *testing.T) {
	limiter := New(3, time.Second)

	for i := 0; i < 3; i++ {
		result := limiter.Allow("user-1")

		if !result.Allowed {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}
}

func TestFixedWindow_RejectsAfterLimit(t *testing.T) {
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

func TestFixedWindow_SeparatesKeys(t *testing.T) {
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

func TestFixedWindow_ResetsAfterWindow(t *testing.T) {
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
		t.Fatal("request should have been allowed after window reset")
	}
}
