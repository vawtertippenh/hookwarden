package storage

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Recorder captures incoming HTTP requests as Events and persists them.
type Recorder struct {
	store Store
}

// NewRecorder creates a Recorder backed by the provided Store.
func NewRecorder(store Store) *Recorder {
	return &Recorder{store: store}
}

// Record reads the request and saves it as an Event, returning the saved event.
func (r *Recorder) Record(req *http.Request, body []byte, signatureOK bool) (*Event, error) {
	headers := make(map[string]string, len(req.Header))
	for k, vals := range req.Header {
		if len(vals) > 0 {
			headers[k] = vals[0]
		}
	}

	ip := req.RemoteAddr
	if fwd := req.Header.Get("X-Forwarded-For"); fwd != "" {
		ip = fwd
	}

	event := &Event{
		ID:          uuid.NewString(),
		ReceivedAt:  time.Now().UTC(),
		Method:      req.Method,
		Path:        req.URL.Path,
		Headers:     headers,
		Body:        body,
		SourceIP:    ip,
		SignatureOK: signatureOK,
	}

	if err := r.store.Save(event); err != nil {
		return nil, fmt.Errorf("recorder: save event: %w", err)
	}
	return event, nil
}

// DrainBody reads and returns all bytes from rc, then closes it.
func DrainBody(rc io.ReadCloser) ([]byte, error) {
	defer rc.Close()
	return io.ReadAll(rc)
}
