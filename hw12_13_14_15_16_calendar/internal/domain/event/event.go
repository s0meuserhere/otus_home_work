package event

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation    = errors.New("event validation error")
	ErrNotFound      = errors.New("event not found")
	ErrDateBusy      = errors.New("date is busy")
	ErrAlreadyExists = errors.New("event already exists")
)

type Event struct {
	id                 uuid.UUID
	title              string
	dateStart          time.Time
	dateEnd            time.Time
	description        string
	userID             uuid.UUID
	notifyShiftSeconds int
}

func NewEvent(
	id uuid.UUID,
	title string,
	dateStart time.Time,
	dateEnd time.Time,
	description string,
	userID uuid.UUID,
	notifyShiftSeconds int,
) (*Event, error) {
	if id == uuid.Nil {
		id = uuid.New()
	}
	if len(title) == 0 {
		return nil, fmt.Errorf("%w: title must not be empty", ErrValidation)
	}

	if dateStart.UTC().Before(time.Now().UTC()) {
		return nil, fmt.Errorf("%w: date start must be in the future", ErrValidation)
	}
	if dateEnd.UTC().Before(time.Now().UTC()) || dateEnd.UTC().Before(dateStart) {
		return nil, fmt.Errorf("%w: date end must be in the future", ErrValidation)
	}

	if len(description) == 0 {
		return nil, fmt.Errorf("%w: description must not be empty", ErrValidation)
	}

	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: userID cannot be empty", ErrValidation)
	}

	if notifyShiftSeconds > 0 {
		duration := time.Duration(notifyShiftSeconds) * time.Second

		if dateStart.After(dateEnd.Add(duration * -1)) {
			return nil, fmt.Errorf("%w: notifyShift must be after date start", ErrValidation)
		}
	}

	return &Event{
		id:                 id,
		title:              title,
		dateStart:          dateStart,
		dateEnd:            dateEnd,
		description:        description,
		userID:             userID,
		notifyShiftSeconds: notifyShiftSeconds,
	}, nil
}

func (e *Event) GetID() uuid.UUID {
	return e.id
}

func (e *Event) GetTitle() string {
	return e.title
}

func (e *Event) GetDateStart() time.Time {
	return e.dateStart
}

func (e *Event) GetDateEnd() time.Time {
	return e.dateEnd
}

func (e *Event) GetDescription() string {
	return e.description
}

func (e *Event) GetUserID() uuid.UUID {
	return e.userID
}

func (e *Event) GetNotifyShiftSeconds() int {
	return e.notifyShiftSeconds
}
