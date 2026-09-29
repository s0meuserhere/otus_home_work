package notify

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/notify"
)

func TestMessageRoundTrip(t *testing.T) {
	t.Parallel()

	eventID := uuid.New()
	userID := uuid.New()
	start := time.Now().Add(time.Hour).UTC().Truncate(time.Second)

	n, err := notify.NewNotify(eventID, "meeting", start, userID)
	if err != nil {
		t.Fatalf("new notify: %v", err)
	}

	data, err := toMessage(*n)
	if err != nil {
		t.Fatalf("to message: %v", err)
	}

	got, err := fromMessage(data)
	if err != nil {
		t.Fatalf("from message: %v", err)
	}

	if got.GetEventID() != eventID || got.GetUserID() != userID || got.GetEventTitle() != "meeting" {
		t.Fatalf("unexpected notify: %+v", got)
	}
	if !got.GetEventDateStart().Equal(start) {
		t.Fatalf("date: got %v, want %v", got.GetEventDateStart(), start)
	}
}

func TestFromMessageInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data string
	}{
		{name: "not json", data: `not json`},
		{name: "wrong field type", data: `{"eventId":123}`},
		// Проверяем, что ошибка домена доходит наружу, сами правила тестируются в domain/notify.
		{name: "domain validation", data: `{"eventTitle":"t","eventDate":"2030-01-01T10:00:00Z"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := fromMessage([]byte(tt.data))
			if !errors.Is(err, notify.ErrValidation) {
				t.Fatalf("err: got %v, want ErrValidation", err)
			}
		})
	}
}
