package api

import (
	"net/http"

	"github.com/yourusername/hookwarden/internal/replay"
	"github.com/yourusername/hookwarden/internal/signature"
	"github.com/yourusername/hookwarden/internal/storage"
)

// RouterConfig holds dependencies needed to build the full HTTP router.
type RouterConfig struct {
	Store     storage.Store
	Replayer  *replay.Replayer
	Validator *signature.Validator
	WebhookPath string
}

// NewRouter constructs and returns the main HTTP mux for hookwarden.
func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// Management API
	eventHandler := NewHandler(cfg.Store)
	mux.HandleFunc("/events", eventHandler.List)
	mux.HandleFunc("/events/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			eventHandler.Get(w, r)
		case http.MethodDelete:
			eventHandler.Delete(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Replay endpoint
	replayHandler := NewReplayHandler(cfg.Store, cfg.Replayer)
	mux.HandleFunc("/events/replay/", replayHandler.ServeHTTP)

	// Stats endpoint
	statsHandler := NewStatsHandler(cfg.Store)
	mux.Handle("/stats", statsHandler)

	// Webhook receiver
	webhookPath := cfg.WebhookPath
	if webhookPath == "" {
		webhookPath = "/webhook"
	}
	mux.Handle(webhookPath, buildWebhookHandler(cfg.Store, cfg.Validator))

	return mux
}

// buildWebhookHandler wires the signature middleware and recorder for a webhook path.
func buildWebhookHandler(store storage.Store, v *signature.Validator) http.Handler {
	recorder := storage.NewRecorder(store)
	var h http.Handler = recorder
	if v != nil {
		h = signature.Middleware(v, h)
	}
	return h
}
