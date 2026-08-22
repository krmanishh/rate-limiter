package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultAlgorithm    = FixedWindow
	defaultStorage      = Memory
	defaultFailMode     = FailClosed
	defaultLimit        = 5
	defaultWindow       = time.Minute
	defaultCapacity     = 10
	defaultRefillRate   = 2.0
	defaultLeakRate     = 2.0
	defaultRedisAddress = "localhost:6379"
	defaultServerPort   = "8080"
	defaultCORSOrigin   = "http://localhost:3000"
)

// ServerConfig holds HTTP server settings, loaded from the environment
// separately from the rate limiter's own configuration.
type ServerConfig struct {
	Port              string
	CORSAllowedOrigin string
}

// LoadServer builds a ServerConfig from environment variables:
//
//	SERVER_PORT         the HTTP server's listen port (default 8080)
//	CORS_ALLOWED_ORIGIN origin allowed to call this API from a browser
//	                    (default http://localhost:3000, the frontend's
//	                    local dev server)
func LoadServer() ServerConfig {
	return ServerConfig{
		Port:              getEnv("SERVER_PORT", defaultServerPort),
		CORSAllowedOrigin: getEnv("CORS_ALLOWED_ORIGIN", defaultCORSOrigin),
	}
}

// Load builds a RateLimitConfig from environment variables, falling back
// to sane defaults for anything unset:
//
//	RATE_LIMIT_ALGORITHM   fixed_window | sliding_log | sliding_counter | token_bucket | leaky_bucket (default fixed_window)
//	RATE_LIMIT_STORAGE     memory | redis (default memory)
//	RATE_LIMIT_FAIL_MODE   closed | open — behavior when Redis errors (default closed)
//	RATE_LIMIT_LIMIT       requests per window (default 5)
//	RATE_LIMIT_WINDOW      window size, e.g. "1m", "30s" (default 1m)
//	RATE_LIMIT_CAPACITY    bucket capacity (default 10)
//	RATE_LIMIT_REFILL_RATE tokens per second (default 2)
//	RATE_LIMIT_LEAK_RATE   requests per second (default 2)
//	REDIS_ADDR             host:port of the Redis server (default localhost:6379)
//	REDIS_PASSWORD         Redis AUTH password (default "", no AUTH)
func Load() (RateLimitConfig, error) {
	algorithm, err := loadAlgorithm()

	if err != nil {
		return RateLimitConfig{}, err
	}

	storage, err := loadStorage()

	if err != nil {
		return RateLimitConfig{}, err
	}

	failMode, err := loadFailMode()

	if err != nil {
		return RateLimitConfig{}, err
	}

	limit, err := getEnvInt("RATE_LIMIT_LIMIT", defaultLimit)

	if err != nil {
		return RateLimitConfig{}, err
	}

	windowSize, err := getEnvDuration("RATE_LIMIT_WINDOW", defaultWindow)

	if err != nil {
		return RateLimitConfig{}, err
	}

	capacity, err := getEnvInt("RATE_LIMIT_CAPACITY", defaultCapacity)

	if err != nil {
		return RateLimitConfig{}, err
	}

	refillRate, err := getEnvFloat("RATE_LIMIT_REFILL_RATE", defaultRefillRate)

	if err != nil {
		return RateLimitConfig{}, err
	}

	leakRate, err := getEnvFloat("RATE_LIMIT_LEAK_RATE", defaultLeakRate)

	if err != nil {
		return RateLimitConfig{}, err
	}

	return RateLimitConfig{
		Algorithm:     algorithm,
		Storage:       storage,
		FailMode:      failMode,
		RedisAddress:  getEnv("REDIS_ADDR", defaultRedisAddress),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		Limit:         limit,
		WindowSize:    windowSize,
		Capacity:      capacity,
		RefillRate:    refillRate,
		LeakRate:      leakRate,
	}, nil
}

func loadAlgorithm() (Algorithm, error) {
	value := Algorithm(getEnv("RATE_LIMIT_ALGORITHM", string(defaultAlgorithm)))

	switch value {
	case FixedWindow, SlidingLog, SlidingCounter, TokenBucket, LeakyBucket:
		return value, nil
	default:
		return "", fmt.Errorf(
			"invalid RATE_LIMIT_ALGORITHM: %q",
			value,
		)
	}
}

func loadStorage() (Storage, error) {
	value := Storage(getEnv("RATE_LIMIT_STORAGE", string(defaultStorage)))

	switch value {
	case Memory, Redis:
		return value, nil
	default:
		return "", fmt.Errorf(
			"invalid RATE_LIMIT_STORAGE: %q",
			value,
		)
	}
}

func loadFailMode() (FailMode, error) {
	value := FailMode(getEnv("RATE_LIMIT_FAIL_MODE", string(defaultFailMode)))

	switch value {
	case FailClosed, FailOpen:
		return value, nil
	default:
		return "", fmt.Errorf(
			"invalid RATE_LIMIT_FAIL_MODE: %q",
			value,
		)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	raw := os.Getenv(key)

	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)

	if err != nil {
		return 0, fmt.Errorf(
			"invalid %s: %w",
			key,
			err,
		)
	}

	return value, nil
}

func getEnvFloat(key string, fallback float64) (float64, error) {
	raw := os.Getenv(key)

	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.ParseFloat(raw, 64)

	if err != nil {
		return 0, fmt.Errorf(
			"invalid %s: %w",
			key,
			err,
		)
	}

	return value, nil
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)

	if raw == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)

	if err != nil {
		return 0, fmt.Errorf(
			"invalid %s: %w",
			key,
			err,
		)
	}

	return value, nil
}
