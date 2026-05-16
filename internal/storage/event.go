package storage

import (
	"time"
)

// Event represents a captured webhook event.
type Event struct {
	ID          string            `json:"id"`
	ReceivedAt  time.Time         `json:"received_at"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
	SourceIP    string            `json:"source_ip"`
	SignatureOK bool              `json:"signature_ok"`
}

// Store defines the interface for persisting and retrieving webhook events.
type Store interface {
	Save(event *Event) error
	Get(id string) (*Event, error)
	List(limit int) ([]*Event, error)
	Delete(id string) error
}
