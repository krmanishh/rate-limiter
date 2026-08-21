package leakybucket

import (
	"context"
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
}
