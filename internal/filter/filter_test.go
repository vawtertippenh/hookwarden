package filter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"hookwarden/internal/filter"
	"hookwarden/internal/storage"
)

func makeEvents() []storage.Event {
	return []storage.Event{
		{ID: "1", Method: "POST", Path: "/hooks/github", Headers: map[string][]string{"X-Source": {"github"}}},
		{ID: "2", Method: "GET", Path: "/hooks/ping", Headers: map[string][]string{"X-Source": {"internal"}}},
		{ID: "3", Method: "POST", Path: "/hooks/stripe", Headers: map[string][]string{"X-Source": {"stripe"}}},
		{ID: "4", Method: "POST", Path: "/hooks/github", Headers: map[string][]string{"X-Source": {"github"}}},
	}
}

func TestApply_NoFilter(t *testing.T) {
	events := makeEvents()
	result := filter.Apply(events, filter.Options{})
	if len(result) != len(events) {
		t.Fatalf("expected %d events, got %d", len(events), len(result))
	}
}

func TestApply_ByMethod(t *testing.T) {
	result := filter.Apply(makeEvents(), filter.Options{Method: "GET"})
	if len(result) != 1 || result[0].ID != "2" {
		t.Fatalf("expected 1 GET event, got %v", result)
	}
}

func TestApply_ByPathPrefix(t *testing.T) {
	result := filter.Apply(makeEvents(), filter.Options{PathPrefix: "/hooks/github"})
	if len(result) != 2 {
		t.Fatalf("expected 2 events, got %d", len(result))
	}
}

func TestApply_ByHeader(t *testing.T) {
	result := filter.Apply(makeEvents(), filter.Options{HeaderKey: "X-Source", HeaderValue: "stripe"})
	if len(result) != 1 || result[0].ID != "3" {
		t.Fatalf("expected 1 stripe event, got %v", result)
	}
}

func TestApply_WithLimit(t *testing.T) {
	result := filter.Apply(makeEvents(), filter.Options{Method: "POST", Limit: 2})
	if len(result) != 2 {
		t.Fatalf("expected 2 events due to limit, got %d", len(result))
	}
}

func TestParseOptions(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?method=post&path=/hooks&header_key=X-Source&header_value=github&limit=5", nil)
	opts := filter.ParseOptions(req)

	if opts.Method != "POST" {
		t.Errorf("expected method POST, got %s", opts.Method)
	}
	if opts.PathPrefix != "/hooks" {
		t.Errorf("expected path /hooks, got %s", opts.PathPrefix)
	}
	if opts.HeaderKey != "X-Source" {
		t.Errorf("expected header_key X-Source, got %s", opts.HeaderKey)
	}
	if opts.Limit != 5 {
		t.Errorf("expected limit 5, got %d", opts.Limit)
	}
}
