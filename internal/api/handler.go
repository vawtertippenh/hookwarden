package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hookwarden/internal/storage"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	store storage.Store
}

// NewHandler creates a new Handler with the given store.
func NewHandler(store storage.Store) *Handler {
	return &Handler{store: store}
}

// ListEvents returns all stored webhook events as JSON.
func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.store.List()
	if err != nil {
		http.Error(w, "failed to list events", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

// GetEvent returns a single webhook event by ID.
func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing event id", http.StatusBadRequest)
		return
	}
	event, err := h.store.Get(id)
	if err != nil {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, event)
}

// DeleteEvent removes a webhook event by ID.
func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing event id", http.StatusBadRequest)
		return
	}
	if err := h.store.Delete(id); err != nil {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HealthCheck returns a simple liveness response.
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC(),
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
