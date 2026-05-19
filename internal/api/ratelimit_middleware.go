package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/yourusername/hookwarden/internal/ratelimit"
)

// RateLimitConfig holds the query-param driven config for the rate-limit admin endpoint.
type RateLimitConfig struct {
	RequestsPerMinute int `json:"requests_per_minute"`
}

// ApplyRateLimit wraps a router with the given rate limiter middleware.
func ApplyRateLimit(limiter *ratelimit.Limiter, next http.Handler) http.Handler {
	return limiter.Middleware(next)
}

// NewRateLimitStatusHandler returns a handler that reports the current rate-limit config.
func NewRateLimitStatusHandler(cfg ratelimit.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Allow overriding RPM via query param for display purposes.
		rpm := cfg.RequestsPerMinute
		if v := r.URL.Query().Get("rpm"); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
				rpm = parsed
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RateLimitConfig{
			RequestsPerMinute: rpm,
		})
	}
}
