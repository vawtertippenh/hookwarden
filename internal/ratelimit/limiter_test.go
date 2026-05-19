package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAllow_WithinLimit(t *testing.T) {
	l := NewLimiter(Config{RequestsPerMinute: 3})
	for i := 0; i < 3; i++ {
		if !l.Allow("client-a") {
			t.Fatalf("expected Allow=true on request %d", i+1)
		}
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	l := NewLimiter(Config{RequestsPerMinute: 2})
	l.Allow("client-b")
	l.Allow("client-b")
	if l.Allow("client-b") {
		t.Fatal("expected Allow=false after limit exceeded")
	}
}

func TestAllow_SeparateKeys(t *testing.T) {
	l := NewLimiter(Config{RequestsPerMinute: 1})
	if !l.Allow("key-1") {
		t.Fatal("expected Allow=true for key-1")
	}
	if !l.Allow("key-2") {
		t.Fatal("expected Allow=true for key-2 (different key)")
	}
}

func TestAllow_WindowReset(t *testing.T) {
	l := NewLimiter(Config{RequestsPerMinute: 1})
	l.Allow("client-c")
	// Manually expire the window
	l.mu.Lock()
	l.buckets["client-c"].windowEnd = time.Now().Add(-time.Second)
	l.mu.Unlock()

	if !l.Allow("client-c") {
		t.Fatal("expected Allow=true after window reset")
	}
}

func TestMiddleware_Allowed(t *testing.T) {
	l := NewLimiter(Config{RequestsPerMinute: 10})
	handler := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_RateLimited(t *testing.T) {
	l := NewLimiter(Config{RequestsPerMinute: 1})
	handler := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	makeReq := func() int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "9.9.9.9:9999"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := makeReq(); code != http.StatusOK {
		t.Fatalf("expected 200 on first request, got %d", code)
	}
	if code := makeReq(); code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on second request, got %d", code)
	}
}
