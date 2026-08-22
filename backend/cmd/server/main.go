package main

import (
	"context"
	"log/slog"
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

const (
	shutdownTimeout = 10 * time.Second

	// Hardening against slow/stalled clients (e.g. slowloris-style
	// resource exhaustion): bound how long a single request can take
	// to send/receive, and how long an idle keep-alive connection is
	// held open.
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 120 * time.Second
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	os.Exit(run())
}

// run contains the full startup/shutdown lifecycle. It returns the
// process exit code rather than calling os.Exit itself, so deferred
// cleanup (stop(), cancel(), closer.Close()) always runs.
func run() int {
	cfg, err := config.Load()

	if err != nil {
		slog.Error("failed to load config", "error", err)
		return 1
	}

	serverCfg := config.LoadServer()

	rateLimiter, closer, err := factory.Create(cfg)

	if err != nil {
		slog.Error("failed to create rate limiter", "error", err)
		return 1
	}

	handler := api.NewHandler(rateLimiter, cfg)

	rateLimitMiddleware := middleware.NewRateLimitMiddleware(
		rateLimiter,
		string(cfg.Algorithm),
		string(cfg.Storage),
	)

	protectedResource := rateLimitMiddleware.Handler(
		http.HandlerFunc(api.ProtectedResource),
	)

	router := api.NewRouter(handler, protectedResource)

	corsRouter := middleware.CORS(serverCfg.CORSAllowedOrigin, router)

	addr := ":" + serverCfg.Port

	server := &http.Server{
		Addr:              addr,
		Handler:           corsRouter,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErr := make(chan error, 1)

	go func() {
		slog.Info(
			"rate limiter server starting",
			"addr", addr,
			"algorithm", cfg.Algorithm,
			"storage", cfg.Storage,
			"cors_allowed_origin", serverCfg.CORSAllowedOrigin,
		)

		serverErr <- server.ListenAndServe()
	}()

	exitCode := 0

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			exitCode = 1
		}

	case <-ctx.Done():
		slog.Info("shutdown signal received, finishing in-flight requests")

		stop()

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("error during server shutdown", "error", err)
			exitCode = 1
		} else {
			slog.Info("server stopped accepting new requests")
		}
	}

	if err := closer.Close(); err != nil {
		slog.Error("error closing rate limiter store", "error", err)
		exitCode = 1
	} else {
		slog.Info("rate limiter store closed")
	}

	slog.Info("exiting", "exit_code", exitCode)

	return exitCode
}
