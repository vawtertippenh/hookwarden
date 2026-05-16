package storage

import (
	"fmt"
	"sync"
)

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu     sync.RWMutex
	events map[string]*Event
	order  []string
}

// NewMemoryStore creates a new MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		events: make(map[string]*Event),
	}
}

// Save persists an event to memory.
func (s *MemoryStore) Save(event *Event) error {
	if event == nil {
		return fmt.Errorf("event must not be nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.events[event.ID]; !exists {
		s.order = append(s.order, event.ID)
	}
	s.events[event.ID] = event
	return nil
}

// Get retrieves an event by ID.
func (s *MemoryStore) Get(id string) (*Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.events[id]
	if !ok {
		return nil, fmt.Errorf("event %q not found", id)
	}
	return e, nil
}

// List returns the most recent events up to limit.
func (s *MemoryStore) List(limit int) ([]*Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.order
	if limit > 0 && limit < len(ids) {
		ids = ids[len(ids)-limit:]
	}
	result := make([]*Event, 0, len(ids))
	for _, id := range ids {
		result = append(result, s.events[id])
	}
	return result, nil
}

// Delete removes an event by ID.
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.events[id]; !ok {
		return fmt.Errorf("event %q not found", id)
	}
	delete(s.events, id)
	for i, oid := range s.order {
		if oid == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return nil
}
