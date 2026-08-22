package slidinglog

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/krmanishh/rate-limiter/internal/store/redisstore"
)

func BenchmarkSlidingLog_Memory(b *testing.B) {
	limiter := New(1<<30, 10*time.Millisecond)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		limiter.Allow("bench-user")
	}
}

func BenchmarkSlidingLog_Redis(b *testing.B) {
	store := redisstore.New("localhost:6379")
	defer store.Close()

	ctx := context.Background()

	if err := store.Ping(ctx); err != nil {
		b.Skipf("redis not available: %v", err)
	}

	limiter := NewRedis(store, 1<<30, 10*time.Millisecond)
	key := fmt.Sprintf("bench-slidinglog-%d", time.Now().UnixNano())

	defer store.Delete(ctx, "rate-limit:slidinglog:"+key)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := limiter.Allow(ctx, key); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
