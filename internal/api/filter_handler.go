package api

import (
	"encoding/json"
	"net/http"

	"hookwarden/internal/filter"
	"hookwarden/internal/storage"
)

// FilterHandler handles GET /events?method=...&path=...&header_key=...&header_value=...&limit=...
type FilterHandler struct {
	store interface {
		List() ([]storage.Event, error)
	}
}

// NewFilterHandler creates a FilterHandler backed by the given store.
func NewFilterHandler(store interface {
	List() ([]storage.Event, error)
}) *FilterHandler {
	return &FilterHandler{store: store}
}

func (h *FilterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	events, err := h.store.List()
	if err != nil {
		http.Error(w, "failed to list events", http.StatusInternalServerError)
		return
	}

	opts := filter.ParseOptions(r)
	result := filter.Apply(events, opts)

	if result == nil {
		result = []storage.Event{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"count":  len(result),
		"events": result,
	})
}
