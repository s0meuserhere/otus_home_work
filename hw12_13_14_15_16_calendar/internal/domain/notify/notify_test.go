package notify_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/notify"
)

func TestNewNotify(t *testing.T) {
	t.Parallel()

	eventID := uuid.New()
	userID := uuid.New()
	future := time.Now().Add(time.Hour)
	past := time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		eventID   uuid.UUID
		title     string
		dateStart time.Time
		userID    uuid.UUID
		wantErr   error
	}{
		{
			name:      "valid",
			eventID:   eventID,
			title:     "meeting",
			dateStart: future,
			userID:    userID,
		},
		{
			// Событие могло начаться, пока уведомление ждало в очереди.
			name:      "event already started",
			eventID:   eventID,
			title:     "meeting",
			dateStart: past,
			userID:    userID,
		},
		{
			name:      "no event id",
			title:     "meeting",
			dateStart: future,
			userID:    userID,
			wantErr:   notify.ErrValidation,
		},
		{
			name:      "no title",
			eventID:   eventID,
			dateStart: future,
			userID:    userID,
			wantErr:   notify.ErrValidation,
		},
		{
			name:    "no date",
			eventID: eventID,
			title:   "meeting",
			userID:  userID,
			wantErr: notify.ErrValidation,
		},
		{
			name:      "no user",
			eventID:   eventID,
			title:     "meeting",
			dateStart: future,
			wantErr:   notify.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			n, err := notify.NewNotify(tt.eventID, tt.title, tt.dateStart, tt.userID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err: got %v, want %v", err, tt.wantErr)
				}

				return
			}
			if err != nil {
				t.Fatalf("new notify: %v", err)
			}

			if n.GetEventID() != tt.eventID || n.GetEventTitle() != tt.title ||
				!n.GetEventDateStart().Equal(tt.dateStart) || n.GetUserID() != tt.userID {
				t.Fatalf("unexpected notify: %+v", n)
			}
		})
	}
}
