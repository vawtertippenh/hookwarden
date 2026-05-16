package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hookwarden/internal/api"
	"github.com/hookwarden/internal/storage"
)

func seedStore(t *testing.T) storage.Store {
	t.Helper()
	s := storage.NewMemoryStore()
	err := s.Save(storage.Event{
		ID:        "evt-1",
		Source:    "github",
		Headers:   map[string][]string{"Content-Type": {"application/json"}},
		Body:      []byte(`{"action":"push"}`),
		ReceivedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	return s
}

func TestListEvents(t *testing.T) {
	h := api.NewHandler(seedStore(t))
	rec := httptest.NewRecorder()
	h.ListEvents(rec, httptest.NewRequest(http.MethodGet, "/events", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var events []storage.Event
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestGetEvent_Found(t *testing.T) {
	h := api.NewHandler(seedStore(t))
	req := httptest.NewRequest(http.MethodGet, "/events/evt-1", nil)
	req.SetPathValue("id", "evt-1")
	rec := httptest.NewRecorder()
	h.GetEvent(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGetEvent_NotFound(t *testing.T) {
	h := api.NewHandler(seedStore(t))
	req := httptest.NewRequest(http.MethodGet, "/events/missing", nil)
	req.SetPathValue("id", "missing")
	rec := httptest.NewRecorder()
	h.GetEvent(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteEvent(t *testing.T) {
	h := api.NewHandler(seedStore(t))
	req := httptest.NewRequest(http.MethodDelete, "/events/evt-1", nil)
	req.SetPathValue("id", "evt-1")
	rec := httptest.NewRecorder()
	h.DeleteEvent(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestHealthCheck(t *testing.T) {
	h := api.NewHandler(storage.NewMemoryStore())
	rec := httptest.NewRecorder()
	h.HealthCheck(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
