package main

import (
	"log"
	"net/http"
	"time"

	"github.com/krmanishh/rate-limiter/internal/api"
	"github.com/krmanishh/rate-limiter/internal/config"
	"github.com/krmanishh/rate-limiter/internal/limiter/factory"
	"github.com/krmanishh/rate-limiter/internal/middleware"
)

func main() {
	cfg := config.RateLimitConfig{
		Algorithm:  config.FixedWindow,
		Limit:      5,
		WindowSize: time.Minute,
	}

	rateLimiter, err := factory.Create(cfg)

	if err != nil {
		log.Fatalf(
			"failed to create rate limiter: %v",
			err,
		)
	}

	handler := api.NewHandler(rateLimiter)

	router := api.NewRouter(handler)

	rateLimitMiddleware := middleware.NewRateLimitMiddleware(
		rateLimiter,
	)

	protectedResource := rateLimitMiddleware.Handler(
		http.HandlerFunc(apiResourceHandler),
	)

	router.Handle(
		"/api/v1/protected-resource",
		protectedResource,
	)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println(
		"rate limiter server running on :8080",
	)

	err = server.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		log.Fatalf(
			"server failed: %v",
			err,
		)
	}
}

func apiResourceHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	w.Write([]byte(`{"message":"protected resource accessed"}`))
}
