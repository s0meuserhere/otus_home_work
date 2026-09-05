package event_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/event"
)

func TestNewEvent(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Date(2030, 6, 15, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)

	tests := []struct {
		name               string
		id                 uuid.UUID
		title              string
		dateStart          time.Time
		dateEnd            time.Time
		description        string
		userID             uuid.UUID
		notifyShiftSeconds int
		wantErr            error
		check              func(t *testing.T, e *event.Event)
	}{
		{
			name:        "ok",
			id:          eventID,
			title:       "meeting",
			dateStart:   start,
			dateEnd:     end,
			description: "desc",
			userID:      userID,
			check: func(t *testing.T, e *event.Event) {
				t.Helper()
				if e.GetID() != eventID {
					t.Fatalf("id: got %s, want %s", e.GetID(), eventID)
				}
				if e.GetTitle() != "meeting" {
					t.Fatalf("title: got %q", e.GetTitle())
				}
				if !e.GetDateStart().Equal(start) {
					t.Fatalf("dateStart: got %v", e.GetDateStart())
				}
				if !e.GetDateEnd().Equal(end) {
					t.Fatalf("dateEnd: got %v", e.GetDateEnd())
				}
				if e.GetDescription() != "desc" {
					t.Fatalf("description: got %q", e.GetDescription())
				}
				if e.GetUserID() != userID {
					t.Fatalf("userID: got %s", e.GetUserID())
				}
				if e.GetNotifyShiftSeconds() != 0 {
					t.Fatalf("notifyShift: got %d", e.GetNotifyShiftSeconds())
				}
			},
		},
		{
			name:        "nil id generates new",
			id:          uuid.Nil,
			title:       "meeting",
			dateStart:   start,
			dateEnd:     end,
			description: "desc",
			userID:      userID,
			check: func(t *testing.T, e *event.Event) {
				t.Helper()
				if e.GetID() == uuid.Nil {
					t.Fatal("expected generated id")
				}
			},
		},
		{
			name:               "ok with notify shift",
			id:                 eventID,
			title:              "meeting",
			dateStart:          start,
			dateEnd:            end,
			description:        "desc",
			userID:             userID,
			notifyShiftSeconds: 3600,
			check: func(t *testing.T, e *event.Event) {
				t.Helper()
				if e.GetNotifyShiftSeconds() != 3600 {
					t.Fatalf("notifyShift: got %d", e.GetNotifyShiftSeconds())
				}
			},
		},
		{
			name:        "empty title",
			id:          eventID,
			title:       "",
			dateStart:   start,
			dateEnd:     end,
			description: "desc",
			userID:      userID,
			wantErr:     event.ErrValidation,
		},
		{
			name:        "date start in past",
			id:          eventID,
			title:       "meeting",
			dateStart:   time.Now().UTC().Add(-time.Hour),
			dateEnd:     end,
			description: "desc",
			userID:      userID,
			wantErr:     event.ErrValidation,
		},
		{
			name:        "date end in past",
			id:          eventID,
			title:       "meeting",
			dateStart:   start,
			dateEnd:     time.Now().UTC().Add(-time.Minute),
			description: "desc",
			userID:      userID,
			wantErr:     event.ErrValidation,
		},
		{
			name:        "date end before start",
			id:          eventID,
			title:       "meeting",
			dateStart:   end,
			dateEnd:     start,
			description: "desc",
			userID:      userID,
			wantErr:     event.ErrValidation,
		},
		{
			name:        "empty description",
			id:          eventID,
			title:       "meeting",
			dateStart:   start,
			dateEnd:     end,
			description: "",
			userID:      userID,
			wantErr:     event.ErrValidation,
		},
		{
			name:        "empty user id",
			id:          eventID,
			title:       "meeting",
			dateStart:   start,
			dateEnd:     end,
			description: "desc",
			userID:      uuid.Nil,
			wantErr:     event.ErrValidation,
		},
		{
			name:               "notify shift too large",
			id:                 eventID,
			title:              "meeting",
			dateStart:          start,
			dateEnd:            end,
			description:        "desc",
			userID:             userID,
			notifyShiftSeconds: int((3 * time.Hour).Seconds()),
			wantErr:            event.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := event.NewEvent(
				tt.id,
				tt.title,
				tt.dateStart,
				tt.dateEnd,
				tt.description,
				tt.userID,
				tt.notifyShiftSeconds,
			)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err: got %v, want %v", err, tt.wantErr)
				}
				if got != nil {
					t.Fatal("expected nil event")
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
