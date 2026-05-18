package replay

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hookwarden/internal/storage"
)

// Result holds the outcome of a replay attempt.
type Result struct {
	EventID    string
	StatusCode int
	Duration   time.Duration
	Err        error
}

// Replayer sends stored webhook events to a target URL.
type Replayer struct {
	store  storage.Store
	client *http.Client
}

// NewReplayer creates a Replayer backed by the given store.
func NewReplayer(store storage.Store, timeout time.Duration) *Replayer {
	return &Replayer{
		store: store,
		client: &http.Client{Timeout: timeout},
	}
}

// Replay fetches event by ID and forwards it to targetURL.
func (r *Replayer) Replay(ctx context.Context, eventID, targetURL string) (*Result, error) {
	event, err := r.store.Get(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("replay: event not found: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, event.Method, targetURL, bytes.NewReader(event.Body))
	if err != nil {
		return nil, fmt.Errorf("replay: build request: %w", err)
	}

	for key, vals := range event.Headers {
		for _, v := range vals {
			req.Header.Add(key, v)
		}
	}
	req.Header.Set("X-Hookwarden-Replay", "true")
	req.Header.Set("X-Hookwarden-Event-ID", event.ID)

	start := time.Now()
	resp, err := r.client.Do(req)
	duration := time.Since(start)
	if err != nil {
		return &Result{EventID: eventID, Duration: duration, Err: err}, nil
	}
	defer resp.Body.Close()

	return &Result{
		EventID:    eventID,
		StatusCode: resp.StatusCode,
		Duration:   duration,
	}, nil
}
