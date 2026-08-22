package api

import (
	"net/http"

	"github.com/krmanishh/rate-limiter/internal/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewRouter owns the application's route table. protectedResource is the
// already middleware-wrapped handler for the rate-limited resource;
// composing that middleware chain remains the caller's responsibility.
// Every route (except /metrics itself) is wrapped with HTTP metrics.
func NewRouter(
	handler *Handler,
	protectedResource http.Handler,
) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle(
		"/health",
		middleware.HTTPMetrics("/health", http.HandlerFunc(healthCheck)),
	)

	mux.Handle(
		"/api/v1/ratelimit/check",
		middleware.HTTPMetrics("/api/v1/ratelimit/check", http.HandlerFunc(handler.CheckRateLimit)),
	)

	mux.Handle(
		"/api/v1/protected-resource",
		middleware.HTTPMetrics("/api/v1/protected-resource", protectedResource),
	)

	mux.Handle(
		"/metrics",
		promhttp.Handler(),
	)

	return mux
}

func healthCheck(
	w http.ResponseWriter,
	_ *http.Request,
) {
	writeJSON(
		w,
		http.StatusOK,
		map[string]string{"status": "ok"},
	)
}

// ProtectedResource is the demo resource guarded by the rate limit
// middleware in cmd/server/main.go.
func ProtectedResource(
	w http.ResponseWriter,
	_ *http.Request,
) {
	writeJSON(
		w,
		http.StatusOK,
		map[string]string{"message": "protected resource accessed"},
	)
}
