package tokenbucket

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
		"rate-limit:tokenbucket:"+key,
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

	if result.RetryAfter < 0 {
		t.Fatalf(
			"expected non-negative retry after, got %d",
			result.RetryAfter,
		)
	}
}

// TestRedisLimiter_RecoversAfterRefill proves tokens actually refill over
// time: once the bucket is drained, waiting long enough for the refill
// rate to add back at least one token should let the next request through.
func TestRedisLimiter_RecoversAfterRefill(t *testing.T) {
	ctx := context.Background()

	redisStore := redisstore.New("localhost:6379")

	defer redisStore.Close()

	if err := redisStore.Ping(ctx); err != nil {
		t.Fatalf("Redis is not available: %v", err)
	}

	limiter := NewRedis(
		redisStore,
		3,
		5, // 5 tokens/sec refill rate
	)

	key := "refill-user-" + time.Now().Format("20060102150405.000000000")

	defer redisStore.Delete(
		ctx,
		"rate-limit:tokenbucket:"+key,
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

	// At 5 tokens/sec, 300ms refills ~1.5 tokens.
	time.Sleep(300 * time.Millisecond)

	result, err := limiter.Allow(ctx, key)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Allowed {
		t.Fatal("request after refill should be allowed")
	}
}

// TestRedisLimiter_Concurrent proves the refill/consume sequence inside
// the Lua script is atomic: firing more requests than the capacity at once
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

	// Negligible refill rate so no extra tokens accrue during the burst.
	limiter := NewRedis(
		redisStore,
		capacity,
		0.0001,
	)

	key := "concurrent-user-" + time.Now().Format("20060102150405.000000000")

	defer redisStore.Delete(
		ctx,
		"rate-limit:tokenbucket:"+key,
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
