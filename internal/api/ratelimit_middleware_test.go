package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/hookwarden/internal/ratelimit"
)

func TestApplyRateLimit_Passthrough(t *testing.T) {
	limiter := ratelimit.NewLimiter(ratelimit.Config{RequestsPerMinute: 10})
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := ApplyRateLimit(limiter, inner)

	req := httptest.NewRequest(http.MethodPost, "/hook", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestApplyRateLimit_Blocked(t *testing.T) {
	limiter := ratelimit.NewLimiter(ratelimit.Config{RequestsPerMinute: 1})
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := ApplyRateLimit(limiter, inner)

	send := func() int {
		req := httptest.NewRequest(http.MethodGet, "/hook", nil)
		req.RemoteAddr = "10.0.0.2:5678"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := send(); code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if code := send(); code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", code)
	}
}

func TestRateLimitStatusHandler_Default(t *testing.T) {
	cfg := ratelimit.Config{RequestsPerMinute: 60}
	h := NewRateLimitStatusHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/admin/ratelimit", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result RateLimitConfig
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.RequestsPerMinute != 60 {
		t.Errorf("expected 60 rpm, got %d", result.RequestsPerMinute)
	}
}

func TestRateLimitStatusHandler_QueryOverride(t *testing.T) {
	cfg := ratelimit.Config{RequestsPerMinute: 60}
	h := NewRateLimitStatusHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/admin/ratelimit?rpm=120", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var result RateLimitConfig
	json.NewDecoder(rec.Body).Decode(&result)
	if result.RequestsPerMinute != 120 {
		t.Errorf("expected 120 rpm, got %d", result.RequestsPerMinute)
	}
}

func TestRateLimitStatusHandler_MethodNotAllowed(t *testing.T) {
	cfg := ratelimit.Config{RequestsPerMinute: 30}
	h := NewRateLimitStatusHandler(cfg)

	req := httptest.NewRequest(http.MethodPost, "/admin/ratelimit", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
