package model

type RateLimitRequest struct {
	Key string `json:"key"`
}

type RateLimitResponse struct {
	Allowed    bool `json:"allowed"`
	Remaining  int  `json:"remaining"`
	RetryAfter int  `json:"retry_after"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
