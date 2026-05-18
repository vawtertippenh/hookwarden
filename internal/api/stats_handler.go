package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourusername/hookwarden/internal/storage"
)

// StatsResponse holds aggregate statistics about stored webhook events.
type StatsResponse struct {
	TotalEvents  int            `json:"total_events"`
	ByMethod     map[string]int `json:"by_method"`
	ByPath       map[string]int `json:"by_path"`
	OldestEvent  *time.Time     `json:"oldest_event,omitempty"`
	NewestEvent  *time.Time     `json:"newest_event,omitempty"`
}

// StatsHandler handles requests for webhook event statistics.
type StatsHandler struct {
	store storage.Store
}

// NewStatsHandler creates a new StatsHandler backed by the given store.
func NewStatsHandler(store storage.Store) *StatsHandler {
	return &StatsHandler{store: store}
}

// ServeHTTP computes and returns aggregate stats over all stored events.
func (h *StatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	events, err := h.store.List()
	if err != nil {
		http.Error(w, "failed to list events", http.StatusInternalServerError)
		return
	}

	stats := StatsResponse{
		TotalEvents: len(events),
		ByMethod:    make(map[string]int),
		ByPath:      make(map[string]int),
	}

	for _, e := range events {
		stats.ByMethod[e.Method]++
		stats.ByPath[e.Path]++

		t := e.ReceivedAt
		if stats.OldestEvent == nil || t.Before(*stats.OldestEvent) {
			stats.OldestEvent = &t
		}
		if stats.NewestEvent == nil || t.After(*stats.NewestEvent) {
			stats.NewestEvent = &t
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
