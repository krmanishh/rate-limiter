package leakybucket

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/krmanishh/rate-limiter/internal/store/redisstore"
)

func TestRedisLimiter(t *testing.T) {
	ctx := context.Background()

	redisStore := redisstore.New("localhost:6379")

	defer redisStore.Close()

	err := redisStore.Ping(ctx)

	if err != nil {
		t.Fatalf(
			"Redis is not available: %v",
			err,
		)
	}

	limiter := NewRedis(
		redisStore,
		3,
		2,
	)

	key := "test-user-" + time.Now().Format("20060102150405.000000000")

	defer redisStore.Delete(
		ctx,
		"rate-limit:leakybucket:"+key,
	)

	for i := int64(1); i <= 3; i++ {
		result, err := limiter.Allow(
			ctx,
			key,
		)

		if err != nil {
			t.Fatalf(
				"unexpected error: %v",
				err,
			)
		}

		if !result.Allowed {
			t.Fatalf(
				"request %d should be allowed",
				i,
			)
		}

		expectedRemaining := int64(3) - i

		if result.Remaining != int(expectedRemaining) {
			t.Fatalf(
				"expected remaining %d, got %d",
				expectedRemaining,
				result.Remaining,
			)
		}
	}

	result, err := limiter.Allow(
		ctx,
		key,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if result.Allowed {
		t.Fatal(
			"fourth request should be rejected",
		)
	}

	if result.Remaining != 0 {
		t.Fatalf(
			"expected remaining 0, got %d",
			result.Remaining,
		)
	}

	if result.RetryAfter <= 0 {
		t.Fatalf(
			"expected a positive RetryAfter, got %d",
			result.RetryAfter,
		)
	}
}

// TestRedisLimiter_RecoversAfterLeak proves the queue actually leaks over
// time: once the bucket is full, waiting long enough for the leak rate to
// drain at least one slot should let the next request through.
func TestRedisLimiter_RecoversAfterLeak(t *testing.T) {
	ctx := context.Background()

	redisStore := redisstore.New("localhost:6379")

	defer redisStore.Close()

	if err := redisStore.Ping(ctx); err != nil {
		t.Fatalf("Redis is not available: %v", err)
	}

	limiter := NewRedis(
		redisStore,
		3,
		5, // 5 requests/sec leak rate
	)

	key := "leak-user-" + time.Now().Format("20060102150405.000000000")

	defer redisStore.Delete(
		ctx,
		"rate-limit:leakybucket:"+key,
	)

	for i := 1; i <= 3; i++ {
		result, err := limiter.Allow(ctx, key)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.Allowed {
			t.Fatalf("request %d should be allowed", i)
		}
	}

	blocked, err := limiter.Allow(ctx, key)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if blocked.Allowed {
		t.Fatal("fourth request should be rejected")
	}

	// At 5 requests/sec, 300ms leaks floor(0.3*5) = 1 slot.
	time.Sleep(300 * time.Millisecond)

	result, err := limiter.Allow(ctx, key)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Allowed {
		t.Fatal("request after leak should be allowed")
	}
}

// TestRedisLimiter_Concurrent proves the leak/enqueue sequence inside the
// Lua script is atomic: firing more requests than the capacity at once
// must still let exactly `capacity` of them through.
func TestRedisLimiter_Concurrent(t *testing.T) {
	ctx := context.Background()

	redisStore := redisstore.New("localhost:6379")

	defer redisStore.Close()

	if err := redisStore.Ping(ctx); err != nil {
		t.Fatalf("Redis is not available: %v", err)
	}

	const capacity = 5
	const attempts = 10

	// Negligible leak rate so the queue doesn't drain during the burst.
	limiter := NewRedis(
		redisStore,
		capacity,
		0.0001,
	)

	key := "concurrent-user-" + time.Now().Format("20060102150405.000000000")

	defer redisStore.Delete(
		ctx,
		"rate-limit:leakybucket:"+key,
	)

	var wg sync.WaitGroup

	var allowedCount int64

	for i := 0; i < attempts; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			result, err := limiter.Allow(ctx, key)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Allowed {
				atomic.AddInt64(&allowedCount, 1)
			}
		}()
	}

	wg.Wait()

	if allowedCount != capacity {
		t.Fatalf(
			"expected exactly %d allowed requests out of %d concurrent attempts, got %d",
			capacity,
			attempts,
			allowedCount,
		)
	}
}
