package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/krmanishh/rate-limiter/internal/config"
	"github.com/krmanishh/rate-limiter/internal/limiter"
	"github.com/krmanishh/rate-limiter/internal/model"
)

// maxCheckRequestBodyBytes caps the /api/v1/ratelimit/check request body.
// The only expected payload is {"key":"..."}; 1 KiB is generous headroom
// while still preventing an oversized or malicious body from being
// buffered in full before decoding fails.
const maxCheckRequestBodyBytes = 1 << 10

type Handler struct {
	limiter limiter.RateLimiter
	cfg     config.RateLimitConfig
}

func NewHandler(rateLimiter limiter.RateLimiter, cfg config.RateLimitConfig) *Handler {
	return &Handler{
		limiter: rateLimiter,
		cfg:     cfg,
	}
}

// GetConfig reports the server's active rate limit configuration, so a
// client (e.g. a dashboard) can display it without separate access to
// the server's environment.
func (h *Handler) GetConfig(
	w http.ResponseWriter,
	_ *http.Request,
) {
	writeJSON(
		w,
		http.StatusOK,
		model.ConfigResponse{
			Algorithm:     string(h.cfg.Algorithm),
			Storage:       string(h.cfg.Storage),
			FailMode:      string(h.cfg.FailMode),
			Limit:         h.cfg.Limit,
			WindowSeconds: int(h.cfg.WindowSize.Seconds()),
			Capacity:      h.cfg.Capacity,
			RefillRate:    h.cfg.RefillRate,
			LeakRate:      h.cfg.LeakRate,
		},
	)
}

func (h *Handler) CheckRateLimit(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			"method not allowed",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxCheckRequestBodyBytes)

	var request model.RateLimitRequest

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	if request.Key == "" {
		writeError(
			w,
			http.StatusBadRequest,
			"key is required",
		)
		return
	}

	result := h.limiter.Allow(request.Key)

	response := model.RateLimitResponse{
		Allowed:    result.Allowed,
		Remaining:  result.Remaining,
		RetryAfter: result.RetryAfter,
	}

	writeJSON(
		w,
		http.StatusOK,
		response,
	)
}

func writeJSON(
	w http.ResponseWriter,
	statusCode int,
	data interface{},
) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to write response", "error", err)
	}
}

func writeError(
	w http.ResponseWriter,
	statusCode int,
	message string,
) {
	writeJSON(
		w,
		statusCode,
		model.ErrorResponse{
			Error: message,
		},
	)
}
