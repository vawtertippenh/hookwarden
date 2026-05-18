package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hookwarden/internal/replay"
	"github.com/hookwarden/internal/storage"
)

// ReplayHandler handles POST /events/{id}/replay requests.
type ReplayHandler struct {
	replayer *replay.Replayer
}

// NewReplayHandler creates a ReplayHandler using the given store.
func NewReplayHandler(store storage.Store) *ReplayHandler {
	return &ReplayHandler{
		replayer: replay.NewReplayer(store, 10*time.Second),
	}
}

type replayRequest struct {
	TargetURL string `json:"target_url"`
}

type replayResponse struct {
	EventID    string        `json:"event_id"`
	StatusCode int           `json:"status_code,omitempty"`
	DurationMs int64         `json:"duration_ms"`
	Error      string        `json:"error,omitempty"`
}

func (h *ReplayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing event id", http.StatusBadRequest)
		return
	}

	var req replayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TargetURL == "" {
		http.Error(w, "invalid request body: target_url required", http.StatusBadRequest)
		return
	}

	result, err := h.replayer.Replay(r.Context(), id, req.TargetURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := replayResponse{
		EventID:    result.EventID,
		StatusCode: result.StatusCode,
		DurationMs: result.Duration.Milliseconds(),
	}
	if result.Err != nil {
		resp.Error = result.Err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
