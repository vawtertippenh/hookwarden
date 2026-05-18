package replay_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hookwarden/internal/replay"
	"github.com/hookwarden/internal/storage"
)

func sampleEvent(id string) storage.Event {
	return storage.Event{
		ID:     id,
		Method: http.MethodPost,
		Path:   "/webhook",
		Headers: http.Header{
			"Content-Type": {"application/json"},
		},
		Body:      []byte(`{"hello":"world"}`),
		ReceivedAt: time.Now(),
	}
}

func TestReplay_Success(t *testing.T) {
	var received []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Hookwarden-Replay") != "true" {
			t.Error("missing replay header")
		}
		w.WriteHeader(http.StatusOK)
		_ = received
	}))
	defer ts.Close()

	store := storage.NewMemoryStore()
	ev := sampleEvent("evt-1")
	_ = store.Save(context.Background(), ev)

	r := replay.NewReplayer(store, 5*time.Second)
	res, err := r.Replay(context.Background(), "evt-1", ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}
	if res.EventID != "evt-1" {
		t.Errorf("expected evt-1, got %s", res.EventID)
	}
}

func TestReplay_EventNotFound(t *testing.T) {
	store := storage.NewMemoryStore()
	r := replay.NewReplayer(store, 5*time.Second)
	_, err := r.Replay(context.Background(), "missing", "http://localhost")
	if err == nil {
		t.Fatal("expected error for missing event")
	}
}

func TestReplay_NetworkError(t *testing.T) {
	store := storage.NewMemoryStore()
	ev := sampleEvent("evt-2")
	_ = store.Save(context.Background(), ev)

	r := replay.NewReplayer(store, 1*time.Second)
	res, err := r.Replay(context.Background(), "evt-2", "http://127.0.0.1:1")
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}
	if res.Err == nil {
		t.Error("expected network error in result")
	}
}
