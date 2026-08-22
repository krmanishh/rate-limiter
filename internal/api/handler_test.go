package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/krmanishh/rate-limiter/internal/limiter/fixedwindow"
)

func TestCheckRateLimit_AllowsRequest(t *testing.T) {
	rateLimiter := fixedwindow.New(
		2,
		time.Minute,
	)

	handler := NewHandler(rateLimiter)

	body := `{"key":"user-1"}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/ratelimit/check",
		strings.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	handler.CheckRateLimit(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			recorder.Code,
		)
	}

	response := recorder.Body.String()

	if !strings.Contains(response, `"allowed":true`) {
		t.Fatalf(
			"expected allowed=true, got %s",
			response,
		)
	}
}

func TestCheckRateLimit_RejectsAfterLimit(t *testing.T) {
	rateLimiter := fixedwindow.New(
		1,
		time.Minute,
	)

	handler := NewHandler(rateLimiter)

	makeRequest := func() *httptest.ResponseRecorder {
		body := `{"key":"user-1"}`

		request := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/ratelimit/check",
			strings.NewReader(body),
		)

		request.Header.Set(
			"Content-Type",
			"application/json",
		)

		recorder := httptest.NewRecorder()

		handler.CheckRateLimit(
			recorder,
			request,
		)

		return recorder
	}

	first := makeRequest()

	if !strings.Contains(
		first.Body.String(),
		`"allowed":true`,
	) {
		t.Fatalf(
			"first request should be allowed: %s",
			first.Body.String(),
		)
	}

	second := makeRequest()

	if !strings.Contains(
		second.Body.String(),
		`"allowed":false`,
	) {
		t.Fatalf(
			"second request should be rejected: %s",
			second.Body.String(),
		)
	}
}

func TestCheckRateLimit_RejectsEmptyKey(t *testing.T) {
	rateLimiter := fixedwindow.New(
		5,
		time.Minute,
	)

	handler := NewHandler(rateLimiter)

	body := `{"key":""}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/ratelimit/check",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.CheckRateLimit(
		recorder,
		request,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status 400, got %d",
			recorder.Code,
		)
	}
}

func TestCheckRateLimit_RejectsInvalidJSON(t *testing.T) {
	rateLimiter := fixedwindow.New(
		5,
		time.Minute,
	)

	handler := NewHandler(rateLimiter)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/ratelimit/check",
		strings.NewReader(`invalid-json`),
	)

	recorder := httptest.NewRecorder()

	handler.CheckRateLimit(
		recorder,
		request,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status 400, got %d",
			recorder.Code,
		)
	}
}

func TestCheckRateLimit_RejectsWrongMethod(t *testing.T) {
	rateLimiter := fixedwindow.New(
		5,
		time.Minute,
	)

	handler := NewHandler(rateLimiter)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/ratelimit/check",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.CheckRateLimit(
		recorder,
		request,
	)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status 405, got %d",
			recorder.Code,
		)
	}
}

func TestCheckRateLimit_RejectsOversizedBody(t *testing.T) {
	rateLimiter := fixedwindow.New(
		5,
		time.Minute,
	)

	handler := NewHandler(rateLimiter)

	// Oversized JSON body: a key far bigger than any real payload
	// should ever need, well past the 1 KiB cap.
	oversizedKey := strings.Repeat("a", 2<<10)
	body := `{"key":"` + oversizedKey + `"}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/ratelimit/check",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.CheckRateLimit(
		recorder,
		request,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status 400 for an oversized body, got %d",
			recorder.Code,
		)
	}
}
