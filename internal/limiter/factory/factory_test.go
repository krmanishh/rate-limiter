package factory

import (
	"testing"
	"time"

	"github.com/krmanishh/rate-limiter/internal/config"
)

func TestCreate_FixedWindow(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:  config.FixedWindow,
		Limit:      5,
		WindowSize: time.Minute,
	}

	limiter, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if limiter == nil {
		t.Fatal("expected limiter, got nil")
	}
}

func TestCreate_SlidingLog(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:  config.SlidingLog,
		Limit:      5,
		WindowSize: time.Minute,
	}

	limiter, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if limiter == nil {
		t.Fatal("expected limiter, got nil")
	}
}

func TestCreate_SlidingCounter(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:  config.SlidingCounter,
		Limit:      5,
		WindowSize: time.Minute,
	}

	limiter, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if limiter == nil {
		t.Fatal("expected limiter, got nil")
	}
}

func TestCreate_TokenBucket(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:  config.TokenBucket,
		Capacity:   5,
		RefillRate: 2,
	}

	limiter, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if limiter == nil {
		t.Fatal("expected limiter, got nil")
	}
}

func TestCreate_LeakyBucket(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm: config.LeakyBucket,
		Capacity:  5,
		LeakRate:  2,
	}

	limiter, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if limiter == nil {
		t.Fatal("expected limiter, got nil")
	}
}

func TestCreate_RejectsUnknownAlgorithm(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm: "unknown",
	}

	limiter, err := Create(cfg)

	if err == nil {
		t.Fatal("expected error for unknown algorithm")
	}

	if limiter != nil {
		t.Fatal("expected nil limiter for invalid algorithm")
	}
}

func TestCreate_RejectsInvalidFixedWindowConfig(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:  config.FixedWindow,
		Limit:      0,
		WindowSize: time.Minute,
	}

	limiter, err := Create(cfg)

	if err == nil {
		t.Fatal("expected validation error")
	}

	if limiter != nil {
		t.Fatal("expected nil limiter")
	}
}

func TestCreate_RejectsInvalidTokenBucketConfig(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:  config.TokenBucket,
		Capacity:   5,
		RefillRate: 0,
	}

	limiter, err := Create(cfg)

	if err == nil {
		t.Fatal("expected validation error")
	}

	if limiter != nil {
		t.Fatal("expected nil limiter")
	}
}
func TestCreate_ReturnedLimiterWorks(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:  config.FixedWindow,
		Limit:      2,
		WindowSize: time.Second,
	}

	limiter, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := limiter.Allow("user-1")

	if !result.Allowed {
		t.Fatal("first request should have been allowed")
	}

	result = limiter.Allow("user-1")

	if !result.Allowed {
		t.Fatal("second request should have been allowed")
	}

	result = limiter.Allow("user-1")

	if result.Allowed {
		t.Fatal("third request should have been rejected")
	}
}

func TestCreate_FixedWindowRedis(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:    config.FixedWindow,
		Storage:      config.Redis,
		RedisAddress: "localhost:6379",
		Limit:        5,
		WindowSize:   time.Minute,
	}

	limiter, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if limiter == nil {
		t.Fatal("expected limiter, got nil")
	}
}
