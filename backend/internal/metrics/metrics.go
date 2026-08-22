// Package metrics defines the Prometheus metrics this service exposes.
// It's the only package that imports the Prometheus client — the
// limiter implementations (FixedWindow, TokenBucket, ...) and the
// storage layer never reference it.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var rateLimitLabels = []string{"algorithm", "storage"}

var (
	// RateLimitRequestsTotal counts every rate limit decision made,
	// regardless of outcome.
	RateLimitRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rate_limit_requests_total",
		Help: "Total number of rate limit checks performed.",
	}, rateLimitLabels)

	// RateLimitAllowedTotal counts requests the limiter allowed.
	RateLimitAllowedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rate_limit_allowed_total",
		Help: "Total number of requests allowed by the rate limiter.",
	}, rateLimitLabels)

	// RateLimitRejectedTotal counts requests the limiter rejected as
	// over the configured limit.
	RateLimitRejectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rate_limit_rejected_total",
		Help: "Total number of requests rejected by the rate limiter.",
	}, rateLimitLabels)

	// RateLimitErrorsTotal counts failures to reach a rate limit
	// decision at all (e.g. a Redis error), as distinct from a
	// legitimate rejection.
	RateLimitErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rate_limit_errors_total",
		Help: "Total number of errors encountered while evaluating the rate limit.",
	}, rateLimitLabels)
)

var httpLabels = []string{"method", "route", "status"}

var (
	// HTTPRequestsTotal counts every HTTP request served.
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, httpLabels)

	// HTTPRequestDuration observes HTTP request latency in seconds.
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: prometheus.DefBuckets,
	}, httpLabels)
)
