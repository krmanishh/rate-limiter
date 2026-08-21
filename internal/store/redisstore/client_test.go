package redisstore

import (
	"context"
	"os"
	"testing"
)

func TestRedisConnection(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")

	if addr == "" {
		addr = "localhost:6379"
	}

	client := New(addr)

	defer client.Close()

	ctx := context.Background()

	err := client.Ping(ctx)

	if err != nil {
		t.Fatalf(
			"failed to connect to Redis: %v",
			err,
		)
	}
}
