package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/event"
)

type Memory struct {
	mu     sync.RWMutex
	events map[uuid.UUID]event.Event
}

func NewMemory() *Memory {
	return &Memory{
		events: make(map[uuid.UUID]event.Event),
	}
}

func (m *Memory) Create(_ context.Context, e event.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.events[e.GetID()]; ok {
		return fmt.Errorf("%w: %s", event.ErrAlreadyExists, e.GetID())
	}

	m.events[e.GetID()] = e

	return nil
}

func (m *Memory) Update(_ context.Context, eventID uuid.UUID, e event.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.events[eventID]; !ok {
		return fmt.Errorf("%w: %s", event.ErrNotFound, eventID)
	}

	m.events[eventID] = e

	return nil
}

func (m *Memory) Delete(_ context.Context, eventID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.events[eventID]; !ok {
		return fmt.Errorf("%w: %s", event.ErrNotFound, eventID)
	}

	delete(m.events, eventID)

	return nil
}

func (m *Memory) ListInRange(_ context.Context, r *event.Range) ([]event.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]event.Event, 0)

	for _, e := range m.events {
		start := e.GetDateStart()
		if !start.Before(r.From) && start.Before(r.To) {
			result = append(result, e)
		}
	}

	return result, nil
}

func (m *Memory) IsDateBusy(
	_ context.Context,
	userID, excludeID uuid.UUID,
	start, end time.Time,
) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for id, existing := range m.events {
		if existing.GetUserID() != userID || id == excludeID {
			continue
		}

		if event.TimeRangesIntersect(start, end, existing.GetDateStart(), existing.GetDateEnd()) {
			return true, nil
		}
	}

	return false, nil
}
