package storage_test

import (
	"testing"
	"time"

	"github.com/yourusername/hookwarden/internal/storage"
)

func sampleEvent(id string) *storage.Event {
	return &storage.Event{
		ID:         id,
		ReceivedAt: time.Now(),
		Method:     "POST",
		Path:       "/hooks",
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       []byte(`{"key":"value"}`),
		SourceIP:   "127.0.0.1",
	}
}

func TestMemoryStore_SaveAndGet(t *testing.T) {
	s := storage.NewMemoryStore()
	ev := sampleEvent("abc123")
	if err := s.Save(ev); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := s.Get("abc123")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != ev.ID {
		t.Errorf("expected ID %q, got %q", ev.ID, got.ID)
	}
}

func TestMemoryStore_GetNotFound(t *testing.T) {
	s := storage.NewMemoryStore()
	_, err := s.Get("missing")
	if err == nil {
		t.Fatal("expected error for missing event")
	}
}

func TestMemoryStore_List(t *testing.T) {
	s := storage.NewMemoryStore()
	for _, id := range []string{"e1", "e2", "e3"} {
		_ = s.Save(sampleEvent(id))
	}
	events, err := s.List(2)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	s := storage.NewMemoryStore()
	_ = s.Save(sampleEvent("del1"))
	if err := s.Delete("del1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := s.Get("del1")
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}

func TestMemoryStore_SaveNil(t *testing.T) {
	s := storage.NewMemoryStore()
	if err := s.Save(nil); err == nil {
		t.Fatal("expected error when saving nil event")
	}
}
