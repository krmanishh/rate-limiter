package api

import "net/http"

func NewRouter(handler *Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"/api/v1/ratelimit/check",
		handler.CheckRateLimit,
	)

	mux.HandleFunc(
		"/health",
		healthCheck,
	)

	mux.HandleFunc(
		"/api/v1/resource",
		resourceHandler,
	)

	return mux
}

func healthCheck(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	w.Write([]byte(`{"status":"ok"}`))
}

func resourceHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	w.Write([]byte(`{"message":"resource accessed successfully"}`))
}
