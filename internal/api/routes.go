package api

import (
	"net/http"

	"github.com/hookwarden/internal/signature"
	"github.com/hookwarden/internal/storage"
)

// RouterConfig holds configuration for building the router.
type RouterConfig struct {
	Store     storage.Store
	Secret    string
	Algorithm string // e.g. "sha256"
}

// NewRouter wires up all routes and returns an http.Handler.
func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()
	h := NewHandler(cfg.Store)

	// Health
	mux.HandleFunc("GET /health", h.HealthCheck)

	// Webhook ingestion — validates HMAC signature then records the event.
	mux.Handle("POST /hooks/{source}", buildWebhookHandler(cfg))

	// Event inspection API
	mux.HandleFunc("GET /events", h.ListEvents)
	mux.HandleFunc("GET /events/{id}", h.GetEvent)
	mux.HandleFunc("DELETE /events/{id}", h.DeleteEvent)

	return mux
}

// buildWebhookHandler constructs the handler for POST /hooks/{source}.
// When a secret is configured it wraps the recorder with HMAC signature
// validation; otherwise requests are accepted without authentication.
func buildWebhookHandler(cfg RouterConfig) http.Handler {
	recorder := storage.NewRecorder(cfg.Store)
	if cfg.Secret == "" {
		return recorder
	}
	algo := cfg.Algorithm
	if algo == "" {
		algo = "sha256"
	}
	v := signature.NewValidator(cfg.Secret, algo)
	return signature.Middleware(v, recorder)
}
