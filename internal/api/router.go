package api

import "net/http"

// NewRouter owns the application's route table. protectedResource is the
// already middleware-wrapped handler for the rate-limited resource;
// composing that middleware chain remains the caller's responsibility.
func NewRouter(
	handler *Handler,
	protectedResource http.Handler,
) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"/health",
		healthCheck,
	)

	mux.HandleFunc(
		"/api/v1/ratelimit/check",
		handler.CheckRateLimit,
	)

	mux.Handle(
		"/api/v1/protected-resource",
		protectedResource,
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
