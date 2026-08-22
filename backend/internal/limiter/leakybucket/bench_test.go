package leakybucket

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/krmanishh/rate-limiter/internal/store/redisstore"
)

// A very large capacity keeps the queue from filling up over the
// course of a benchmark run, so we're measuring the steady-state
// "allowed" path rather than a mix of allowed/rejected outcomes.
const benchCapacity = math.MaxInt32

func BenchmarkLeakyBucket_Memory(b *testing.B) {
	limiter := New(benchCapacity, 1000)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		limiter.Allow("bench-user")
	}
}

func BenchmarkLeakyBucket_Redis(b *testing.B) {
	store := redisstore.New("localhost:6379")
	defer store.Close()

	ctx := context.Background()

	if err := store.Ping(ctx); err != nil {
		b.Skipf("redis not available: %v", err)
	}

	limiter := NewRedis(store, benchCapacity, 1000)
	key := fmt.Sprintf("bench-leakybucket-%d", time.Now().UnixNano())

	defer store.Delete(ctx, "rate-limit:leakybucket:"+key)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := limiter.Allow(ctx, key); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
