package notify

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/event"
)

var ErrValidation = errors.New("notify validation error")

type Notify struct {
	eventID        uuid.UUID
	eventTitle     string
	eventDateStart time.Time
	userID         uuid.UUID
}

func NewNotify(eventID uuid.UUID, eventTitle string, eventDateStart time.Time, userID uuid.UUID) (*Notify, error) {
	if eventID == uuid.Nil {
		return nil, fmt.Errorf("%w: eventID is required", ErrValidation)
	}

	if len(eventTitle) == 0 {
		return nil, fmt.Errorf("%w: eventTitle is required", ErrValidation)
	}

	if eventDateStart.IsZero() {
		return nil, fmt.Errorf("%w: eventDateStart is required", ErrValidation)
	}

	if eventDateStart.UTC().Before(time.Now().UTC()) {
		return nil, fmt.Errorf("%w: event date start must be in the future", ErrValidation)
	}

	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: userID is required", ErrValidation)
	}

	return &Notify{
		eventID:        eventID,
		eventTitle:     eventTitle,
		eventDateStart: eventDateStart,
		userID:         userID,
	}, nil
}

func NewNotifyFromEvent(event event.Event) (*Notify, error) {
	return NewNotify(
		event.GetID(),
		event.GetTitle(),
		event.GetDateStart(),
		event.GetUserID(),
	)
}
