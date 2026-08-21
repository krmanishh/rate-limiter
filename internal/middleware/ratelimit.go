package middleware

import (
	"net/http"
	"strconv"

	"github.com/krmanishh/rate-limiter/internal/limiter"
)

type RateLimitMiddleware struct {
	limiter limiter.RateLimiter
}

func NewRateLimitMiddleware(
	rateLimiter limiter.RateLimiter,
) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: rateLimiter,
	}
}

func (m *RateLimitMiddleware) Handler(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		key := getRateLimitKey(r)

		result := m.limiter.Allow(key)

		setRateLimitHeaders(w, result)

		if !result.Allowed {
			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusTooManyRequests)

			w.Write([]byte(`{"error":"rate limit exceeded"}`))

			return
		}

		next.ServeHTTP(w, r)
	})
}

func getRateLimitKey(r *http.Request) string {
	apiKey := r.Header.Get("X-API-Key")

	if apiKey != "" {
		return "api-key:" + apiKey
	}

	return "ip:" + r.RemoteAddr
}

func setRateLimitHeaders(
	w http.ResponseWriter,
	result limiter.Result,
) {
	w.Header().Set(
		"X-RateLimit-Remaining",
		strconv.Itoa(result.Remaining),
	)

	if result.RetryAfter > 0 {
		w.Header().Set(
			"Retry-After",
			strconv.Itoa(result.RetryAfter),
		)
	}
}
