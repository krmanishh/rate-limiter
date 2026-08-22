package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krmanishh/rate-limiter/internal/api"
	"github.com/krmanishh/rate-limiter/internal/config"
	"github.com/krmanishh/rate-limiter/internal/limiter/factory"
	"github.com/krmanishh/rate-limiter/internal/middleware"
)

const shutdownTimeout = 10 * time.Second

func main() {
	os.Exit(run())
}

// run contains the full startup/shutdown lifecycle. It returns the
// process exit code rather than calling os.Exit itself, so deferred
// cleanup (stop(), cancel(), closer.Close()) always runs.
func run() int {
	cfg, err := config.Load()

	if err != nil {
		log.Fatalf(
			"failed to load config: %v",
			err,
		)
	}

	serverCfg := config.LoadServer()

	rateLimiter, closer, err := factory.Create(cfg)

	if err != nil {
		log.Fatalf(
			"failed to create rate limiter: %v",
			err,
		)
	}

	handler := api.NewHandler(rateLimiter)

	rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimiter)

	protectedResource := rateLimitMiddleware.Handler(
		http.HandlerFunc(api.ProtectedResource),
	)

	router := api.NewRouter(handler, protectedResource)

	addr := ":" + serverCfg.Port

	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErr := make(chan error, 1)

	go func() {
		log.Printf(
			"rate limiter server running on %s (algorithm=%s storage=%s)",
			addr,
			cfg.Algorithm,
			cfg.Storage,
		)

		serverErr <- server.ListenAndServe()
	}()

	exitCode := 0

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Printf("server failed: %v", err)
			exitCode = 1
		}

	case <-ctx.Done():
		log.Println("shutdown signal received, finishing in-flight requests")

		stop()

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("error during server shutdown: %v", err)
			exitCode = 1
		} else {
			log.Println("server stopped accepting new requests")
		}
	}

	if err := closer.Close(); err != nil {
		log.Printf("error closing rate limiter store: %v", err)
		exitCode = 1
	} else {
		log.Println("rate limiter store closed")
	}

	log.Println("exiting")

	return exitCode
}
