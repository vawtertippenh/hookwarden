package storage_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/yourusername/hookwarden/internal/storage"
)

func TestRecorder_Record(t *testing.T) {
	store := storage.NewMemoryStore()
	rec := storage.NewRecorder(store)

	body := []byte(`{"event":"push"}`)
	req, _ := http.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "10.0.0.1:9000"

	ev, err := rec.Record(req, body, true)
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	if ev.ID == "" {
		t.Error("expected non-empty event ID")
	}
	if ev.Method != http.MethodPost {
		t.Errorf("expected POST, got %q", ev.Method)
	}
	if !ev.SignatureOK {
		t.Error("expected SignatureOK to be true")
	}

	fetched, err := store.Get(ev.ID)
	if err != nil {
		t.Fatalf("Get after Record: %v", err)
	}
	if string(fetched.Body) != string(body) {
		t.Errorf("body mismatch: got %q", fetched.Body)
	}
}

func TestDrainBody(t *testing.T) {
	rc := io.NopCloser(strings.NewReader("hello world"))
	data, err := storage.DrainBody(rc)
	if err != nil {
		t.Fatalf("DrainBody: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("unexpected data: %q", data)
	}
}
