package adapter_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/krmanishh/rate-limiter/internal/limiter/adapter"
	"github.com/krmanishh/rate-limiter/internal/limiter/fixedwindow"
	"github.com/krmanishh/rate-limiter/internal/store/redisstore"
)

// This file exercises what actually happens when the real Redis
// container goes away mid-traffic — not a mock, the real container
// docker-compose.yml defines. That means it stops shared infrastructure
// other tests/packages may be using concurrently, so it's opt-in only:
//
//	REDIS_FAILURE_TEST=1 go test ./internal/limiter/adapter/... -run Failure -v
//
// It's never run by plain `go test ./...` or CI.
const redisFailureContainer = "rate-limiter-redis"

func requireRedisFailureTestOptIn(t *testing.T) {
	t.Helper()

	if os.Getenv("REDIS_FAILURE_TEST") == "" {
		t.Skip("skipping: set REDIS_FAILURE_TEST=1 to run (stops/starts the real rate-limiter-redis container)")
	}

	if err := exec.Command("docker", "inspect", redisFailureContainer).Run(); err != nil {
		t.Skipf("docker container %q not available: %v", redisFailureContainer, err)
	}
}

func stopRedisContainer(t *testing.T) {
	t.Helper()

	if out, err := exec.Command("docker", "stop", redisFailureContainer).CombinedOutput(); err != nil {
		t.Fatalf("failed to stop %s: %v: %s", redisFailureContainer, err, out)
	}

	// Give the container a moment to actually stop accepting connections.
	time.Sleep(500 * time.Millisecond)
}

func startRedisContainer(t *testing.T, store *redisstore.Client) {
	t.Helper()

	if out, err := exec.Command("docker", "start", redisFailureContainer).CombinedOutput(); err != nil {
		t.Fatalf("failed to start %s: %v: %s", redisFailureContainer, err, out)
	}

	deadline := time.Now().Add(15 * time.Second)

	for time.Now().Before(deadline) {
		if err := store.Ping(context.Background()); err == nil {
			return
		}

		time.Sleep(200 * time.Millisecond)
	}

	t.Fatalf("redis did not become reachable again within the deadline")
}

func TestRedisFailure_FailClosed(t *testing.T) {
	requireRedisFailureTestOptIn(t)

	store := redisstore.New("localhost:6379")
	defer store.Close()

	if err := store.Ping(context.Background()); err != nil {
		t.Fatalf("redis is not available before the test even starts: %v", err)
	}

	key := fmt.Sprintf("failure-test-closed-%d", time.Now().UnixNano())

	rateLimiter := adapter.NewRedis(
		fixedwindow.NewRedis(store, 5, time.Minute),
		"fixed_window",
		5,
		false, // fail closed
	)

	// Sanity: works while Redis is up.
	if result := rateLimiter.Allow(key); !result.Allowed {
		t.Fatal("expected request to be allowed while redis is up")
	}

	stopRedisContainer(t)
	defer startRedisContainer(t, store)

	result := rateLimiter.Allow(key)

	if result.Allowed {
		t.Fatal("expected fail-closed to reject the request while redis is down")
	}

	if result.RetryAfter <= 0 {
		t.Fatalf("expected a positive RetryAfter hint, got %d", result.RetryAfter)
	}

	startRedisContainer(t, store)

	result = rateLimiter.Allow(key)

	if !result.Allowed {
		t.Fatal("expected the rate limiter to work normally again once redis recovers")
	}
}

func TestRedisFailure_FailOpen(t *testing.T) {
	requireRedisFailureTestOptIn(t)

	store := redisstore.New("localhost:6379")
	defer store.Close()

	if err := store.Ping(context.Background()); err != nil {
		t.Fatalf("redis is not available before the test even starts: %v", err)
	}

	key := fmt.Sprintf("failure-test-open-%d", time.Now().UnixNano())

	rateLimiter := adapter.NewRedis(
		fixedwindow.NewRedis(store, 5, time.Minute),
		"fixed_window",
		5,
		true, // fail open
	)

	if result := rateLimiter.Allow(key); !result.Allowed {
		t.Fatal("expected request to be allowed while redis is up")
	}

	stopRedisContainer(t)
	defer startRedisContainer(t, store)

	result := rateLimiter.Allow(key)

	if !result.Allowed {
		t.Fatal("expected fail-open to allow the request while redis is down")
	}

	startRedisContainer(t, store)

	result = rateLimiter.Allow(key)

	if !result.Allowed {
		t.Fatal("expected the rate limiter to work normally again once redis recovers")
	}
}
