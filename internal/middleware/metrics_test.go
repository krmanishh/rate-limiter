package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/krmanishh/rate-limiter/internal/limiter/fixedwindow"
	"github.com/krmanishh/rate-limiter/internal/metrics"
)

func TestRateLimitMiddleware_RecordsAllowedMetrics(t *testing.T) {
	rateLimiter := fixedwindow.New(2, time.Minute)

	mw := NewRateLimitMiddleware(rateLimiter, "fixed_window", "memory")

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.Handler(next)

	requestsBefore := testutil.ToFloat64(metrics.RateLimitRequestsTotal.WithLabelValues("fixed_window", "memory"))
	allowedBefore := testutil.ToFloat64(metrics.RateLimitAllowedTotal.WithLabelValues("fixed_window", "memory"))
	rejectedBefore := testutil.ToFloat64(metrics.RateLimitRejectedTotal.WithLabelValues("fixed_window", "memory"))

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	request.Header.Set("X-API-Key", "metrics-allowed-user")

	handler.ServeHTTP(httptest.NewRecorder(), request)

	requestsAfter := testutil.ToFloat64(metrics.RateLimitRequestsTotal.WithLabelValues("fixed_window", "memory"))
	allowedAfter := testutil.ToFloat64(metrics.RateLimitAllowedTotal.WithLabelValues("fixed_window", "memory"))
	rejectedAfter := testutil.ToFloat64(metrics.RateLimitRejectedTotal.WithLabelValues("fixed_window", "memory"))

	if requestsAfter-requestsBefore != 1 {
		t.Fatalf("expected rate_limit_requests_total to increase by 1, got %v -> %v", requestsBefore, requestsAfter)
	}

	if allowedAfter-allowedBefore != 1 {
		t.Fatalf("expected rate_limit_allowed_total to increase by 1, got %v -> %v", allowedBefore, allowedAfter)
	}

	if rejectedAfter != rejectedBefore {
		t.Fatalf("expected rate_limit_rejected_total to stay unchanged, got %v -> %v", rejectedBefore, rejectedAfter)
	}
}

func TestRateLimitMiddleware_RecordsRejectedMetrics(t *testing.T) {
	rateLimiter := fixedwindow.New(1, time.Minute)

	mw := NewRateLimitMiddleware(rateLimiter, "fixed_window", "memory")

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.Handler(next)

	makeRequest := func() {
		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.Header.Set("X-API-Key", "metrics-rejected-user")

		handler.ServeHTTP(httptest.NewRecorder(), request)
	}

	// Exhaust the limit without measuring, then measure exactly one
	// more (rejected) request in isolation.
	makeRequest()

	requestsBefore := testutil.ToFloat64(metrics.RateLimitRequestsTotal.WithLabelValues("fixed_window", "memory"))
	allowedBefore := testutil.ToFloat64(metrics.RateLimitAllowedTotal.WithLabelValues("fixed_window", "memory"))
	rejectedBefore := testutil.ToFloat64(metrics.RateLimitRejectedTotal.WithLabelValues("fixed_window", "memory"))

	makeRequest()

	requestsAfter := testutil.ToFloat64(metrics.RateLimitRequestsTotal.WithLabelValues("fixed_window", "memory"))
	allowedAfter := testutil.ToFloat64(metrics.RateLimitAllowedTotal.WithLabelValues("fixed_window", "memory"))
	rejectedAfter := testutil.ToFloat64(metrics.RateLimitRejectedTotal.WithLabelValues("fixed_window", "memory"))

	if requestsAfter-requestsBefore != 1 {
		t.Fatalf("expected rate_limit_requests_total to increase by 1, got %v -> %v", requestsBefore, requestsAfter)
	}

	if rejectedAfter-rejectedBefore != 1 {
		t.Fatalf("expected rate_limit_rejected_total to increase by 1, got %v -> %v", rejectedBefore, rejectedAfter)
	}

	if allowedAfter != allowedBefore {
		t.Fatalf("expected rate_limit_allowed_total to stay unchanged, got %v -> %v", allowedBefore, allowedAfter)
	}
}
