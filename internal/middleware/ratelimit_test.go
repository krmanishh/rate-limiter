package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/krmanishh/rate-limiter/internal/limiter/fixedwindow"
)

func TestRateLimitMiddleware_AllowsRequest(t *testing.T) {
	rateLimiter := fixedwindow.New(
		2,
		time.Minute,
	)

	middleware := NewRateLimitMiddleware(rateLimiter)

	next := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)

	handler := middleware.Handler(next)

	request := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	request.Header.Set(
		"X-API-Key",
		"user-1",
	)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d",
			recorder.Code,
		)
	}

	if recorder.Header().Get(
		"X-RateLimit-Remaining",
	) == "" {
		t.Fatal(
			"expected X-RateLimit-Remaining header",
		)
	}
}

func TestRateLimitMiddleware_RejectsRequest(t *testing.T) {
	rateLimiter := fixedwindow.New(
		1,
		time.Minute,
	)

	middleware := NewRateLimitMiddleware(rateLimiter)

	nextCalled := 0

	next := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			nextCalled++
			w.WriteHeader(http.StatusOK)
		},
	)

	handler := middleware.Handler(next)

	makeRequest := func() *httptest.ResponseRecorder {
		request := httptest.NewRequest(
			http.MethodGet,
			"/test",
			nil,
		)

		request.Header.Set(
			"X-API-Key",
			"user-1",
		)

		recorder := httptest.NewRecorder()

		handler.ServeHTTP(
			recorder,
			request,
		)

		return recorder
	}

	// First request should be allowed.
	first := makeRequest()

	if first.Code != http.StatusOK {
		t.Fatalf(
			"expected first request to return 200, got %d",
			first.Code,
		)
	}

	if nextCalled != 1 {
		t.Fatalf(
			"expected next handler to be called once, got %d",
			nextCalled,
		)
	}

	// Second request should be rejected.
	second := makeRequest()

	if second.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected second request to return 429, got %d",
			second.Code,
		)
	}

	// The second request must NOT reach the next handler.
	if nextCalled != 1 {
		t.Fatalf(
			"expected next handler to still be called only once, got %d",
			nextCalled,
		)
	}

	if second.Header().Get(
		"X-RateLimit-Remaining",
	) != "0" {
		t.Fatal(
			"expected remaining to be 0",
		)
	}

	if second.Header().Get(
		"X-RateLimit-Limit",
	) != "1" {
		t.Fatalf(
			"expected X-RateLimit-Limit 1, got %q",
			second.Header().Get("X-RateLimit-Limit"),
		)
	}
}

func TestRateLimitMiddleware_SeparatesAPIKeys(t *testing.T) {
	rateLimiter := fixedwindow.New(
		1,
		time.Minute,
	)

	middleware := NewRateLimitMiddleware(rateLimiter)

	next := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)

	handler := middleware.Handler(next)

	makeRequest := func(
		apiKey string,
	) *httptest.ResponseRecorder {
		request := httptest.NewRequest(
			http.MethodGet,
			"/test",
			nil,
		)

		request.Header.Set(
			"X-API-Key",
			apiKey,
		)

		recorder := httptest.NewRecorder()

		handler.ServeHTTP(
			recorder,
			request,
		)

		return recorder
	}

	user1First := makeRequest("user-1")

	if user1First.Code != http.StatusOK {
		t.Fatal("user-1 first request should succeed")
	}

	user1Second := makeRequest("user-1")

	if user1Second.Code != http.StatusTooManyRequests {
		t.Fatal("user-1 second request should be rejected")
	}

	user2First := makeRequest("user-2")

	if user2First.Code != http.StatusOK {
		t.Fatal("user-2 should have its own rate limit")
	}
}
