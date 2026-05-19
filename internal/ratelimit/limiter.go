package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// Config holds rate limiter configuration.
type Config struct {
	RequestsPerMinute int
}

// entry tracks request count and window start for a single key.
type entry struct {
	count     int
	windowEnd time.Time
}

// Limiter is a simple in-memory per-IP rate limiter.
type Limiter struct {
	mu      sync.Mutex
	cfg     Config
	buckets map[string]*entry
}

// NewLimiter creates a new Limiter with the given config.
func NewLimiter(cfg Config) *Limiter {
	if cfg.RequestsPerMinute <= 0 {
		cfg.RequestsPerMinute = 60
	}
	return &Limiter{
		cfg:     cfg,
		buckets: make(map[string]*entry),
	}
}

// Allow returns true if the key is within the rate limit.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	e, ok := l.buckets[key]
	if !ok || now.After(e.windowEnd) {
		l.buckets[key] = &entry{
			count:     1,
			windowEnd: now.Add(time.Minute),
		}
		return true
	}

	if e.count >= l.cfg.RequestsPerMinute {
		return false
	}
	e.count++
	return true
}

// Middleware returns an http.Handler that enforces the rate limit by remote IP.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !l.Allow(ip) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP extracts the client IP from the request, preferring X-Forwarded-For.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	return r.RemoteAddr
}
