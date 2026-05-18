package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hookwarden/internal/api"
	"github.com/hookwarden/internal/storage"
)

func replaySeedStore(t *testing.T) storage.Store {
	t.Helper()
	store := storage.NewMemoryStore()
	ev := storage.Event{
		ID:     "evt-replay",
		Method: http.MethodPost,
		Path:   "/hook",
		Headers: http.Header{"Content-Type": {"application/json"}},
		Body:       []byte(`{"x":1}`),
		ReceivedAt: time.Now(),
	}
	if err := store.Save(context.Background(), ev); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestReplayHandler_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	store := replaySeedStore(t)
	h := api.NewReplayHandler(store)

	body, _ := json.Marshal(map[string]string{"target_url": ts.URL})
	req := httptest.NewRequest(http.MethodPost, "/events/evt-replay/replay", bytes.NewReader(body))
	req.SetPathValue("id", "evt-replay")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp["event_id"] != "evt-replay" {
		t.Errorf("unexpected event_id: %v", resp["event_id"])
	}
}

func TestReplayHandler_MissingTargetURL(t *testing.T) {
	store := replaySeedStore(t)
	h := api.NewReplayHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/events/evt-replay/replay", bytes.NewReader([]byte(`{}`)))
	req.SetPathValue("id", "evt-replay")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestReplayHandler_EventNotFound(t *testing.T) {
	store := storage.NewMemoryStore()
	h := api.NewReplayHandler(store)

	body, _ := json.Marshal(map[string]string{"target_url": "http://localhost"})
	req := httptest.NewRequest(http.MethodPost, "/events/ghost/replay", bytes.NewReader(body))
	req.SetPathValue("id", "ghost")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
