package adapter

import (
	"context"
	"errors"
	"testing"

	"github.com/krmanishh/rate-limiter/internal/limiter"
)

type stubLimiter struct {
	result limiter.Result
	err    error
}

func (s stubLimiter) Allow(_ context.Context, _ string) (limiter.Result, error) {
	return s.result, s.err
}

func TestRedisAdapter_PassesThroughSuccess(t *testing.T) {
	stub := stubLimiter{result: limiter.Result{Allowed: true, Remaining: 3, Limit: 5}}

	a := NewRedis(stub, "fixed_window", 5, false)

	result := a.Allow("key")

	if !result.Allowed || result.Remaining != 3 || result.Limit != 5 {
		t.Fatalf("expected the underlying result to pass through unchanged, got %+v", result)
	}
}

func TestRedisAdapter_FailClosedRejectsOnError(t *testing.T) {
	stub := stubLimiter{err: errors.New("simulated redis failure")}

	a := NewRedis(stub, "fixed_window", 5, false)

	result := a.Allow("key")

	if result.Allowed {
		t.Fatal("expected fail-closed to reject the request on a redis error")
	}

	if result.Limit != 5 {
		t.Fatalf("expected Limit to still be populated, got %d", result.Limit)
	}

	if result.RetryAfter <= 0 {
		t.Fatalf("expected a positive RetryAfter hint, got %d", result.RetryAfter)
	}
}

func TestRedisAdapter_FailOpenAllowsOnError(t *testing.T) {
	stub := stubLimiter{err: errors.New("simulated redis failure")}

	a := NewRedis(stub, "fixed_window", 5, true)

	result := a.Allow("key")

	if !result.Allowed {
		t.Fatal("expected fail-open to allow the request on a redis error")
	}

	if result.Limit != 5 {
		t.Fatalf("expected Limit to still be populated, got %d", result.Limit)
	}
}
