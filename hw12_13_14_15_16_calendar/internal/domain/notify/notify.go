package notify

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrValidation = errors.New("notify validation error")

// Status - результат обработки уведомления рассыльщиком.
type Status string

const StatusSent Status = "sent"

type Notify struct {
	eventID        uuid.UUID
	eventTitle     string
	eventDateStart time.Time
	userID         uuid.UUID
}

// NewNotify допускает прошедшую дату, потому что событие могло начаться, пока уведомление ждало в очереди.
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

func (n *Notify) GetEventID() uuid.UUID {
	return n.eventID
}

func (n *Notify) GetEventTitle() string {
	return n.eventTitle
}

func (n *Notify) GetEventDateStart() time.Time {
	return n.eventDateStart
}

func (n *Notify) GetUserID() uuid.UUID {
	return n.userID
}
