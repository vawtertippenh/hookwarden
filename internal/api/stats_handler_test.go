package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/hookwarden/internal/storage"
)

func statsSeedStore(t *testing.T) storage.Store {
	t.Helper()
	store := storage.NewMemoryStore()

	base := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	events := []storage.Event{
		{ID: "1", Method: "POST", Path: "/hooks/github", ReceivedAt: base},
		{ID: "2", Method: "POST", Path: "/hooks/stripe", ReceivedAt: base.Add(time.Hour)},
		{ID: "3", Method: "GET", Path: "/hooks/github", ReceivedAt: base.Add(2 * time.Hour)},
	}
	for _, e := range events {
		if err := store.Save(e); err != nil {
			t.Fatalf("seed store: %v", err)
		}
	}
	return store
}

func TestStatsHandler_Empty(t *testing.T) {
	h := NewStatsHandler(storage.NewMemoryStore())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stats", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var stats StatsResponse
	if err := json.NewDecoder(rec.Body).Decode(&stats); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if stats.TotalEvents != 0 {
		t.Errorf("expected 0 events, got %d", stats.TotalEvents)
	}
	if stats.OldestEvent != nil || stats.NewestEvent != nil {
		t.Error("expected nil oldest/newest for empty store")
	}
}

func TestStatsHandler_WithEvents(t *testing.T) {
	h := NewStatsHandler(statsSeedStore(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stats", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var stats StatsResponse
	if err := json.NewDecoder(rec.Body).Decode(&stats); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if stats.TotalEvents != 3 {
		t.Errorf("expected 3 total events, got %d", stats.TotalEvents)
	}
	if stats.ByMethod["POST"] != 2 {
		t.Errorf("expected 2 POST events, got %d", stats.ByMethod["POST"])
	}
	if stats.ByMethod["GET"] != 1 {
		t.Errorf("expected 1 GET event, got %d", stats.ByMethod["GET"])
	}
	if stats.ByPath["/hooks/github"] != 2 {
		t.Errorf("expected 2 /hooks/github events, got %d", stats.ByPath["/hooks/github"])
	}
	if stats.OldestEvent == nil || stats.NewestEvent == nil {
		t.Fatal("expected non-nil oldest and newest")
	}
	if !stats.OldestEvent.Before(*stats.NewestEvent) {
		t.Error("oldest should be before newest")
	}
}
