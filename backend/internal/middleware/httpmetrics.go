package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/krmanishh/rate-limiter/internal/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// HTTPMetrics wraps next with http_requests_total and
// http_request_duration_seconds, labeled by method, the route it was
// registered under (not the raw request path), and response status.
func HTTPMetrics(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		start := time.Now()

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(recorder, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(recorder.status)

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, route, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, route, status).Observe(duration)
	})
}
