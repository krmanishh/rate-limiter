package slidinglog

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
		time.Minute,
	)

	key := "test-user-" + time.Now().Format("20060102150405.000000000")

	defer redisStore.Delete(
		ctx,
		"rate-limit:slidinglog:"+key,
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

	if result.RetryAfter < 0 {
		t.Fatalf(
			"expected non-negative retry after, got %d",
			result.RetryAfter,
		)
	}
}

// TestRedisLimiter_RecoversAfterWindow proves the sliding window actually
// slides: once the oldest logged request ages out of the window, a new
// request should be allowed again without waiting for a hard reset.
func TestRedisLimiter_RecoversAfterWindow(t *testing.T) {
	ctx := context.Background()

	redisStore := redisstore.New("localhost:6379")

	defer redisStore.Close()

	if err := redisStore.Ping(ctx); err != nil {
		t.Fatalf("Redis is not available: %v", err)
	}

	limiter := NewRedis(
		redisStore,
		3,
		time.Second,
	)

	key := "recover-user-" + time.Now().Format("20060102150405.000000000")

	defer redisStore.Delete(
		ctx,
		"rate-limit:slidinglog:"+key,
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

	time.Sleep(1100 * time.Millisecond)

	result, err := limiter.Allow(ctx, key)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Allowed {
		t.Fatal("request after the window elapses should be allowed")
	}
}

// TestRedisLimiter_Concurrent proves the ZADD/ZCARD/ZREMRANGEBYSCORE
// sequence inside the Lua script is atomic: firing more requests than the
// limit at once must still let exactly `limit` of them through.
func TestRedisLimiter_Concurrent(t *testing.T) {
	ctx := context.Background()

	redisStore := redisstore.New("localhost:6379")

	defer redisStore.Close()

	if err := redisStore.Ping(ctx); err != nil {
		t.Fatalf("Redis is not available: %v", err)
	}

	const limit = 5
	const attempts = 10

	limiter := NewRedis(
		redisStore,
		limit,
		time.Minute,
	)

	key := "concurrent-user-" + time.Now().Format("20060102150405.000000000")

	defer redisStore.Delete(
		ctx,
		"rate-limit:slidinglog:"+key,
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

	if allowedCount != limit {
		t.Fatalf(
			"expected exactly %d allowed requests out of %d concurrent attempts, got %d",
			limit,
			attempts,
			allowedCount,
		)
	}
}
