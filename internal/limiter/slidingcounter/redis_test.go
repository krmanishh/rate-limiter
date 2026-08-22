package slidingcounter

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
		"rate-limit:slidingcounter:"+key,
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

	if result.Remaining != 0 {
		t.Fatalf(
			"expected remaining 0, got %d",
			result.Remaining,
		)
	}
}

// TestRedisLimiter_WindowTransition exercises the current/previous window
// weighting: a full window rejects further requests, rolling into the next
// window lets requests through again as the previous count's weight decays,
// and once more than two full windows have elapsed the previous count's
// influence is gone entirely.
func TestRedisLimiter_WindowTransition(t *testing.T) {
	ctx := context.Background()

	redisStore := redisstore.New("localhost:6379")

	defer redisStore.Close()

	if err := redisStore.Ping(ctx); err != nil {
		t.Fatalf("Redis is not available: %v", err)
	}

	const limit = 5

	window := 2 * time.Second

	limiter := NewRedis(
		redisStore,
		limit,
		window,
	)

	key := "transition-user-" + time.Now().Format("20060102150405.000000000")

	defer redisStore.Delete(
		ctx,
		"rate-limit:slidingcounter:"+key,
	)

	// Fill the current window completely.
	for i := 1; i <= limit; i++ {
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
		t.Fatal("request should be rejected once the window is full")
	}

	if blocked.Remaining != 0 {
		t.Fatalf("expected remaining 0, got %d", blocked.Remaining)
	}

	// Cross into the next window: the full window becomes "previous" and
	// its weight decays as elapsed time within the new window grows, so a
	// request shortly after rollover should be allowed again.
	time.Sleep(window + 300*time.Millisecond)

	afterTransition, err := limiter.Allow(ctx, key)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !afterTransition.Allowed {
		t.Fatal("request should be allowed shortly after the window rolls over")
	}

	// Well past two full windows, the previous window's influence is gone
	// entirely and the limiter behaves like a fresh window again.
	time.Sleep(window * 5 / 2)

	for i := 1; i <= limit; i++ {
		result, err := limiter.Allow(ctx, key)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.Allowed {
			t.Fatalf("fresh-window request %d should be allowed", i)
		}
	}
}

// TestRedisLimiter_Concurrent proves the HMGET/HSET read-modify-write
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
		"rate-limit:slidingcounter:"+key,
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
