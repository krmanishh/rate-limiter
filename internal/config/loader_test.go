package config

import (
	"testing"
	"time"
)

func clearEnv(t *testing.T) {
	keys := []string{
		"RATE_LIMIT_ALGORITHM",
		"RATE_LIMIT_STORAGE",
		"RATE_LIMIT_FAIL_MODE",
		"RATE_LIMIT_LIMIT",
		"RATE_LIMIT_WINDOW",
		"RATE_LIMIT_CAPACITY",
		"RATE_LIMIT_REFILL_RATE",
		"RATE_LIMIT_LEAK_RATE",
		"REDIS_ADDR",
		"REDIS_PASSWORD",
		"SERVER_PORT",
	}

	for _, key := range keys {
		t.Setenv(key, "")
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Algorithm != FixedWindow {
		t.Fatalf("expected default algorithm %q, got %q", FixedWindow, cfg.Algorithm)
	}

	if cfg.Storage != Memory {
		t.Fatalf("expected default storage %q, got %q", Memory, cfg.Storage)
	}

	if cfg.Limit != 5 {
		t.Fatalf("expected default limit 5, got %d", cfg.Limit)
	}

	if cfg.WindowSize != time.Minute {
		t.Fatalf("expected default window 1m, got %s", cfg.WindowSize)
	}

	if cfg.Capacity != 10 {
		t.Fatalf("expected default capacity 10, got %d", cfg.Capacity)
	}

	if cfg.RefillRate != 2 {
		t.Fatalf("expected default refill rate 2, got %v", cfg.RefillRate)
	}

	if cfg.LeakRate != 2 {
		t.Fatalf("expected default leak rate 2, got %v", cfg.LeakRate)
	}

	if cfg.RedisAddress != "localhost:6379" {
		t.Fatalf("expected default redis address, got %q", cfg.RedisAddress)
	}

	if cfg.FailMode != FailClosed {
		t.Fatalf("expected default fail mode %q, got %q", FailClosed, cfg.FailMode)
	}

	if cfg.RedisPassword != "" {
		t.Fatalf("expected default redis password to be empty, got %q", cfg.RedisPassword)
	}
}

func TestLoad_OverridesFromEnv(t *testing.T) {
	clearEnv(t)

	t.Setenv("RATE_LIMIT_ALGORITHM", "token_bucket")
	t.Setenv("RATE_LIMIT_STORAGE", "redis")
	t.Setenv("RATE_LIMIT_LIMIT", "42")
	t.Setenv("RATE_LIMIT_WINDOW", "30s")
	t.Setenv("RATE_LIMIT_CAPACITY", "20")
	t.Setenv("RATE_LIMIT_REFILL_RATE", "3.5")
	t.Setenv("RATE_LIMIT_LEAK_RATE", "1.5")
	t.Setenv("REDIS_ADDR", "redis-host:6380")
	t.Setenv("REDIS_PASSWORD", "s3cret")
	t.Setenv("RATE_LIMIT_FAIL_MODE", "open")

	cfg, err := Load()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Algorithm != TokenBucket {
		t.Fatalf("expected algorithm token_bucket, got %q", cfg.Algorithm)
	}

	if cfg.Storage != Redis {
		t.Fatalf("expected storage redis, got %q", cfg.Storage)
	}

	if cfg.Limit != 42 {
		t.Fatalf("expected limit 42, got %d", cfg.Limit)
	}

	if cfg.WindowSize != 30*time.Second {
		t.Fatalf("expected window 30s, got %s", cfg.WindowSize)
	}

	if cfg.Capacity != 20 {
		t.Fatalf("expected capacity 20, got %d", cfg.Capacity)
	}

	if cfg.RefillRate != 3.5 {
		t.Fatalf("expected refill rate 3.5, got %v", cfg.RefillRate)
	}

	if cfg.LeakRate != 1.5 {
		t.Fatalf("expected leak rate 1.5, got %v", cfg.LeakRate)
	}

	if cfg.RedisAddress != "redis-host:6380" {
		t.Fatalf("expected redis address override, got %q", cfg.RedisAddress)
	}

	if cfg.FailMode != FailOpen {
		t.Fatalf("expected fail mode override %q, got %q", FailOpen, cfg.FailMode)
	}

	if cfg.RedisPassword != "s3cret" {
		t.Fatalf("expected redis password override, got %q", cfg.RedisPassword)
	}
}

func TestLoad_RejectsInvalidFailMode(t *testing.T) {
	clearEnv(t)
	t.Setenv("RATE_LIMIT_FAIL_MODE", "not-a-real-fail-mode")

	_, err := Load()

	if err == nil {
		t.Fatal("expected error for invalid fail mode")
	}
}

func TestLoad_RejectsInvalidAlgorithm(t *testing.T) {
	clearEnv(t)
	t.Setenv("RATE_LIMIT_ALGORITHM", "not-a-real-algorithm")

	_, err := Load()

	if err == nil {
		t.Fatal("expected error for invalid algorithm")
	}
}

func TestLoad_RejectsInvalidStorage(t *testing.T) {
	clearEnv(t)
	t.Setenv("RATE_LIMIT_STORAGE", "not-a-real-storage")

	_, err := Load()

	if err == nil {
		t.Fatal("expected error for invalid storage")
	}
}

func TestLoad_RejectsNonNumericLimit(t *testing.T) {
	clearEnv(t)
	t.Setenv("RATE_LIMIT_LIMIT", "not-a-number")

	_, err := Load()

	if err == nil {
		t.Fatal("expected error for non-numeric limit")
	}
}

func TestLoad_RejectsInvalidWindow(t *testing.T) {
	clearEnv(t)
	t.Setenv("RATE_LIMIT_WINDOW", "not-a-duration")

	_, err := Load()

	if err == nil {
		t.Fatal("expected error for invalid window duration")
	}
}

func TestLoad_RejectsNonNumericCapacity(t *testing.T) {
	clearEnv(t)
	t.Setenv("RATE_LIMIT_CAPACITY", "not-a-number")

	_, err := Load()

	if err == nil {
		t.Fatal("expected error for non-numeric capacity")
	}
}

func TestLoad_RejectsNonNumericRefillRate(t *testing.T) {
	clearEnv(t)
	t.Setenv("RATE_LIMIT_REFILL_RATE", "not-a-number")

	_, err := Load()

	if err == nil {
		t.Fatal("expected error for non-numeric refill rate")
	}
}

func TestLoad_RejectsNonNumericLeakRate(t *testing.T) {
	clearEnv(t)
	t.Setenv("RATE_LIMIT_LEAK_RATE", "not-a-number")

	_, err := Load()

	if err == nil {
		t.Fatal("expected error for non-numeric leak rate")
	}
}

func TestLoadServer_Default(t *testing.T) {
	clearEnv(t)

	cfg := LoadServer()

	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %q", cfg.Port)
	}
}

func TestLoadServer_Override(t *testing.T) {
	clearEnv(t)
	t.Setenv("SERVER_PORT", "9090")

	cfg := LoadServer()

	if cfg.Port != "9090" {
		t.Fatalf("expected port override 9090, got %q", cfg.Port)
	}
}
