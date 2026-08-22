package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	"github.com/krmanishh/rate-limiter/internal/limiter"
	"github.com/krmanishh/rate-limiter/internal/metrics"
)

type RateLimitMiddleware struct {
	limiter   limiter.RateLimiter
	algorithm string
	storage   string
}

// NewRateLimitMiddleware wraps rateLimiter for use as HTTP middleware.
// algorithm and storage are only used as metric/log labels — the
// limiter itself never sees or reports them.
func NewRateLimitMiddleware(
	rateLimiter limiter.RateLimiter,
	algorithm string,
	storage string,
) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter:   rateLimiter,
		algorithm: algorithm,
		storage:   storage,
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

		metrics.RateLimitRequestsTotal.WithLabelValues(m.algorithm, m.storage).Inc()

		if result.Allowed {
			metrics.RateLimitAllowedTotal.WithLabelValues(m.algorithm, m.storage).Inc()
		} else {
			metrics.RateLimitRejectedTotal.WithLabelValues(m.algorithm, m.storage).Inc()

			slog.Info(
				"request rate limited",
				"key_hash", hashKey(key),
				"algorithm", m.algorithm,
				"storage", m.storage,
				"remaining", result.Remaining,
			)
		}

		setRateLimitHeaders(w, result)

		if !result.Allowed {
			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusTooManyRequests)

			if _, err := w.Write([]byte(`{"error":"rate limit exceeded"}`)); err != nil {
				slog.Error("failed to write rate limit response", "error", err)
			}

			return
		}

		next.ServeHTTP(w, r)
	})
}

// getRateLimitKey identifies the caller for rate-limiting purposes: the
// X-API-Key header if present, otherwise the connecting IP.
//
// This deliberately does NOT consult X-Forwarded-For or similar
// headers. Trusting a forwarded-for header requires knowing this
// service sits behind a specific proxy that sets it correctly and
// strips any client-supplied value first — something this service has
// no way to verify on its own. Without that guarantee, any client
// could set the header themselves and rate-limit as someone else (or
// evade limits entirely by spoofing a fresh IP on every request). If
// this service is ever deployed behind a trusted reverse proxy, that
// proxy's real client IP extraction should replace this function, not
// extend it to blindly trust a header.
func getRateLimitKey(r *http.Request) string {
	apiKey := r.Header.Get("X-API-Key")

	if apiKey != "" {
		return "api-key:" + apiKey
	}

	return "ip:" + clientIP(r)
}

// clientIP returns the request's remote address without its ephemeral
// port. r.RemoteAddr is "ip:port", and the port is different on every
// connection — keying on the raw value would mean the IP-based fallback
// never actually rate-limits anything, since no two requests would ever
// share a key.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		return r.RemoteAddr
	}

	return host
}

// hashKey avoids writing raw rate-limit keys — which may embed a
// caller-supplied API key — into logs. A truncated SHA-256 is enough to
// correlate repeated hits from the same key across log lines without
// exposing the value itself.
func hashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])[:12]
}

func setRateLimitHeaders(
	w http.ResponseWriter,
	result limiter.Result,
) {
	w.Header().Set(
		"X-RateLimit-Limit",
		strconv.Itoa(result.Limit),
	)

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
