package slidingcounter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/krmanishh/rate-limiter/internal/store/redisstore"
)

func BenchmarkSlidingCounter_Memory(b *testing.B) {
	limiter := New(1<<30, time.Hour)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		limiter.Allow("bench-user")
	}
}

func BenchmarkSlidingCounter_Redis(b *testing.B) {
	store := redisstore.New("localhost:6379")
	defer store.Close()

	ctx := context.Background()

	if err := store.Ping(ctx); err != nil {
		b.Skipf("redis not available: %v", err)
	}

	limiter := NewRedis(store, 1<<30, time.Hour)
	key := fmt.Sprintf("bench-slidingcounter-%d", time.Now().UnixNano())

	defer store.Delete(ctx, "rate-limit:slidingcounter:"+key)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := limiter.Allow(ctx, key); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
