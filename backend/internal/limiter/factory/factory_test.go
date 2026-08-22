package factory

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/krmanishh/rate-limiter/internal/config"
	"github.com/krmanishh/rate-limiter/internal/limiter"
)

func TestCreate_FixedWindow(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:  config.FixedWindow,
		Limit:      5,
		WindowSize: time.Minute,
	}

	limiter, closer, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer closer.Close()

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

	limiter, closer, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer closer.Close()

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

	limiter, closer, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer closer.Close()

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

	limiter, closer, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer closer.Close()

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

	limiter, closer, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer closer.Close()

	if limiter == nil {
		t.Fatal("expected limiter, got nil")
	}
}

func TestCreate_RejectsUnknownAlgorithm(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm: "unknown",
	}

	limiter, closer, err := Create(cfg)

	if err == nil {
		t.Fatal("expected error for unknown algorithm")
	}

	if limiter != nil {
		t.Fatal("expected nil limiter for invalid algorithm")
	}

	if closer != nil {
		t.Fatal("expected nil closer for invalid algorithm")
	}
}

func TestCreate_RejectsInvalidFixedWindowConfig(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:  config.FixedWindow,
		Limit:      0,
		WindowSize: time.Minute,
	}

	limiter, closer, err := Create(cfg)

	if err == nil {
		t.Fatal("expected validation error")
	}

	if limiter != nil {
		t.Fatal("expected nil limiter")
	}

	if closer != nil {
		t.Fatal("expected nil closer")
	}
}

func TestCreate_RejectsInvalidTokenBucketConfig(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:  config.TokenBucket,
		Capacity:   5,
		RefillRate: 0,
	}

	limiter, closer, err := Create(cfg)

	if err == nil {
		t.Fatal("expected validation error")
	}

	if limiter != nil {
		t.Fatal("expected nil limiter")
	}

	if closer != nil {
		t.Fatal("expected nil closer")
	}
}

func TestCreate_ReturnedLimiterWorks(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:  config.FixedWindow,
		Limit:      2,
		WindowSize: time.Second,
	}

	limiter, closer, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer closer.Close()

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

func TestCreate_MemoryStorageReturnsNoopCloser(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:  config.FixedWindow,
		Limit:      5,
		WindowSize: time.Minute,
	}

	_, closer, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if closer == nil {
		t.Fatal("expected a non-nil closer even for memory storage")
	}

	if err := closer.Close(); err != nil {
		t.Fatalf("expected memory storage closer to be a no-op, got error: %v", err)
	}
}

func TestCreate_FixedWindowRedis(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:    config.FixedWindow,
		Storage:      config.Redis,
		RedisAddress: "localhost:6379",
		Limit:        3,
		WindowSize:   time.Minute,
	}

	limiter, closer, err := Create(cfg)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	defer closer.Close()

	if limiter == nil {
		t.Fatal("expected limiter, got nil")
	}

	key := fmt.Sprintf("factory-test-user-%d", time.Now().UnixNano())

	first := limiter.Allow(key)

	if !first.Allowed {
		t.Fatal("first request should be allowed")
	}

	second := limiter.Allow(key)

	if !second.Allowed {
		t.Fatal("second request should be allowed")
	}

	third := limiter.Allow(key)

	if !third.Allowed {
		t.Fatal("third request should be allowed")
	}

	fourth := limiter.Allow(key)

	if fourth.Allowed {
		t.Fatal("fourth request should be rejected")
	}
}

func TestCreate_SlidingLogRedis(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:    config.SlidingLog,
		Storage:      config.Redis,
		RedisAddress: "localhost:6379",
		Limit:        3,
		WindowSize:   time.Minute,
	}

	limiter, closer, err := Create(cfg)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	defer closer.Close()

	key := fmt.Sprintf("factory-test-user-%d", time.Now().UnixNano())

	for i := 0; i < 3; i++ {
		result := limiter.Allow(key)

		if !result.Allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	fourth := limiter.Allow(key)

	if fourth.Allowed {
		t.Fatal("fourth request should be rejected")
	}
}

func TestCreate_SlidingCounterRedis(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:    config.SlidingCounter,
		Storage:      config.Redis,
		RedisAddress: "localhost:6379",
		Limit:        3,
		WindowSize:   time.Minute,
	}

	limiter, closer, err := Create(cfg)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	defer closer.Close()

	key := fmt.Sprintf("factory-test-user-%d", time.Now().UnixNano())

	for i := 0; i < 3; i++ {
		result := limiter.Allow(key)

		if !result.Allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	fourth := limiter.Allow(key)

	if fourth.Allowed {
		t.Fatal("fourth request should be rejected")
	}
}

func TestCreate_TokenBucketRedis(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:    config.TokenBucket,
		Storage:      config.Redis,
		RedisAddress: "localhost:6379",
		Capacity:     3,
		RefillRate:   2,
	}

	limiter, closer, err := Create(cfg)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	defer closer.Close()

	key := fmt.Sprintf("factory-test-user-%d", time.Now().UnixNano())

	for i := 0; i < 3; i++ {
		result := limiter.Allow(key)

		if !result.Allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	fourth := limiter.Allow(key)

	if fourth.Allowed {
		t.Fatal("fourth request should be rejected")
	}
}

func TestCreate_LeakyBucketRedis(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:    config.LeakyBucket,
		Storage:      config.Redis,
		RedisAddress: "localhost:6379",
		Capacity:     3,
		LeakRate:     2,
	}

	limiter, closer, err := Create(cfg)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	defer closer.Close()

	key := fmt.Sprintf("factory-test-user-%d", time.Now().UnixNano())

	for i := 0; i < 3; i++ {
		result := limiter.Allow(key)

		if !result.Allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	fourth := limiter.Allow(key)

	if fourth.Allowed {
		t.Fatal("fourth request should be rejected")
	}
}

func TestCreate_RejectsMissingRedisAddress(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:  config.FixedWindow,
		Storage:    config.Redis,
		Limit:      5,
		WindowSize: time.Minute,
	}

	limiter, closer, err := Create(cfg)

	if err == nil {
		t.Fatal("expected error for missing redis address")
	}

	if limiter != nil {
		t.Fatal("expected nil limiter")
	}

	if closer != nil {
		t.Fatal("expected nil closer")
	}
}

func TestCreate_RedisStorageClosesCleanly(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:    config.FixedWindow,
		Storage:      config.Redis,
		RedisAddress: "localhost:6379",
		Limit:        5,
		WindowSize:   time.Minute,
	}

	_, closer, err := Create(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if closer == nil {
		t.Fatal("expected a non-nil closer for redis storage")
	}

	if err := closer.Close(); err != nil {
		t.Fatalf("expected redis client to close cleanly, got error: %v", err)
	}
}

// TestCreate_DistributedAcrossInstances is the actual reason Redis-backed
// rate limiting exists: it proves the limit is enforced on the shared key,
// not per process. Three independently created limiters — each with its
// own Redis client, standing in for three separate app instances — all
// hammer the same key concurrently. If each instance kept its own count
// (as a memory limiter would), a limit of 5 would let through 5 x 3 = 15
// requests total instead of 5.
func TestCreate_DistributedAcrossInstances(t *testing.T) {
	cfg := config.RateLimitConfig{
		Algorithm:    config.FixedWindow,
		Storage:      config.Redis,
		RedisAddress: "localhost:6379",
		Limit:        5,
		WindowSize:   time.Minute,
	}

	const instanceCount = 3
	const requestsPerInstance = 10 // 3 x 10 = 30 concurrent attempts against a limit of 5

	instances := make([]limiter.RateLimiter, instanceCount)

	for i := 0; i < instanceCount; i++ {
		rateLimiter, closer, err := Create(cfg)

		if err != nil {
			t.Fatalf("failed to create instance %d: %v", i, err)
		}

		t.Cleanup(func() {
			closer.Close()
		})

		instances[i] = rateLimiter
	}

	key := fmt.Sprintf("distributed-test-%d", time.Now().UnixNano())

	var wg sync.WaitGroup

	var allowedCount int64

	for i := 0; i < instanceCount; i++ {
		instance := instances[i]

		for j := 0; j < requestsPerInstance; j++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				if instance.Allow(key).Allowed {
					atomic.AddInt64(&allowedCount, 1)
				}
			}()
		}
	}

	wg.Wait()

	if allowedCount != int64(cfg.Limit) {
		t.Fatalf(
			"expected exactly %d allowed across %d simulated instances (not %d x %d = %d), got %d",
			cfg.Limit,
			instanceCount,
			instanceCount,
			requestsPerInstance,
			instanceCount*requestsPerInstance,
			allowedCount,
		)
	}
}
