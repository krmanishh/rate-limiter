package api_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/krmanishh/rate-limiter/internal/api"
	"github.com/krmanishh/rate-limiter/internal/config"
	"github.com/krmanishh/rate-limiter/internal/limiter/factory"
	"github.com/krmanishh/rate-limiter/internal/metrics"
	"github.com/krmanishh/rate-limiter/internal/middleware"
	"github.com/krmanishh/rate-limiter/internal/store/redisstore"
)

// newIntegrationServer wires up the same request path main.go builds —
// router -> middleware -> RateLimiter interface -> concrete limiter — and
// serves it over a real HTTP listener so tests exercise the whole stack,
// not just individual components in isolation.
func newIntegrationServer(t *testing.T, cfg config.RateLimitConfig) *httptest.Server {
	t.Helper()

	rateLimiter, closer, err := factory.Create(cfg)

	if err != nil {
		t.Fatalf("failed to create rate limiter: %v", err)
	}

	t.Cleanup(func() {
		if err := closer.Close(); err != nil {
			t.Errorf("failed to close rate limiter store: %v", err)
		}
	})

	handler := api.NewHandler(rateLimiter)

	rateLimitMiddleware := middleware.NewRateLimitMiddleware(
		rateLimiter,
		string(cfg.Algorithm),
		string(cfg.Storage),
	)

	protectedResource := rateLimitMiddleware.Handler(
		http.HandlerFunc(api.ProtectedResource),
	)

	router := api.NewRouter(handler, protectedResource)

	server := httptest.NewServer(router)

	t.Cleanup(server.Close)

	return server
}

func TestIntegration_HealthCheck(t *testing.T) {
	server := newIntegrationServer(t, config.RateLimitConfig{
		Algorithm:  config.FixedWindow,
		Limit:      5,
		WindowSize: time.Minute,
	})

	resp, err := http.Get(server.URL + "/health")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// exerciseProtectedResource fires `limit` allowed requests followed by one
// rejected request against /api/v1/protected-resource, asserting the full
// HTTP request -> router -> middleware -> RateLimiter -> response chain
// behaves correctly end to end.
func exerciseProtectedResource(t *testing.T, server *httptest.Server, apiKey string, limit int) {
	t.Helper()

	client := server.Client()

	for i := 1; i <= limit; i++ {
		req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/protected-resource", nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		req.Header.Set("X-API-Key", apiKey)

		resp, err := client.Do(req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, resp.StatusCode)
		}
	}

	req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/protected-resource", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req.Header.Set("X-API-Key", apiKey)

	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 once the limit is exceeded, got %d", resp.StatusCode)
	}

	if got := resp.Header.Get("X-RateLimit-Limit"); got != fmt.Sprintf("%d", limit) {
		t.Fatalf("expected X-RateLimit-Limit %d, got %q", limit, got)
	}

	if got := resp.Header.Get("X-RateLimit-Remaining"); got != "0" {
		t.Fatalf("expected X-RateLimit-Remaining 0, got %q", got)
	}

	if resp.Header.Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header on a 429 response")
	}
}

func TestIntegration_ProtectedResource_Memory(t *testing.T) {
	server := newIntegrationServer(t, config.RateLimitConfig{
		Algorithm:  config.FixedWindow,
		Storage:    config.Memory,
		Limit:      5,
		WindowSize: time.Minute,
	})

	apiKey := fmt.Sprintf("integration-user-%d", time.Now().UnixNano())

	exerciseProtectedResource(t, server, apiKey, 5)
}

func TestIntegration_HTTPMetrics(t *testing.T) {
	server := newIntegrationServer(t, config.RateLimitConfig{
		Algorithm:  config.FixedWindow,
		Storage:    config.Memory,
		Limit:      1,
		WindowSize: time.Minute,
	})

	client := server.Client()

	healthBefore := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, "/health", "200"))

	resp, err := client.Get(server.URL + "/health")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	healthAfter := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, "/health", "200"))

	if healthAfter-healthBefore != 1 {
		t.Fatalf(
			"expected http_requests_total{method=GET,route=/health,status=200} to increase by 1, got %v -> %v",
			healthBefore,
			healthAfter,
		)
	}

	apiKey := fmt.Sprintf("http-metrics-user-%d", time.Now().UnixNano())
	route := "/api/v1/protected-resource"

	allowedBefore := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, route, "200"))
	rejectedBefore := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, route, "429"))

	for i, wantStatus := range []int{http.StatusOK, http.StatusTooManyRequests} {
		req, err := http.NewRequest(http.MethodGet, server.URL+route, nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		req.Header.Set("X-API-Key", apiKey)

		resp, err := client.Do(req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resp.Body.Close()

		if resp.StatusCode != wantStatus {
			t.Fatalf("request %d: expected %d, got %d", i+1, wantStatus, resp.StatusCode)
		}
	}

	allowedAfter := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, route, "200"))
	rejectedAfter := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, route, "429"))

	if allowedAfter-allowedBefore != 1 {
		t.Fatalf(
			"expected http_requests_total{route=%s,status=200} to increase by 1, got %v -> %v",
			route,
			allowedBefore,
			allowedAfter,
		)
	}

	if rejectedAfter-rejectedBefore != 1 {
		t.Fatalf(
			"expected http_requests_total{route=%s,status=429} to increase by 1, got %v -> %v",
			route,
			rejectedBefore,
			rejectedAfter,
		)
	}

	// /metrics itself should serve Prometheus-compatible plaintext
	// exposition containing the counters this test just moved.
	metricsResp, err := client.Get(server.URL + "/metrics")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer metricsResp.Body.Close()

	if metricsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /metrics, got %d", metricsResp.StatusCode)
	}

	body, err := io.ReadAll(metricsResp.Body)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(string(body), "http_requests_total") {
		t.Fatal("expected /metrics output to contain http_requests_total")
	}

	if !strings.Contains(string(body), "rate_limit_requests_total") {
		t.Fatal("expected /metrics output to contain rate_limit_requests_total")
	}
}

func TestIntegration_ProtectedResource_Redis(t *testing.T) {
	redisStore := redisstore.New("localhost:6379")

	defer redisStore.Close()

	if err := redisStore.Ping(context.Background()); err != nil {
		t.Fatalf("Redis is not available: %v", err)
	}

	server := newIntegrationServer(t, config.RateLimitConfig{
		Algorithm:    config.FixedWindow,
		Storage:      config.Redis,
		RedisAddress: "localhost:6379",
		Limit:        5,
		WindowSize:   time.Minute,
	})

	apiKey := fmt.Sprintf("integration-user-%d", time.Now().UnixNano())

	t.Cleanup(func() {
		redisStore.Delete(context.Background(), "rate-limit:fixed:api-key:"+apiKey)
	})

	exerciseProtectedResource(t, server, apiKey, 5)
}
