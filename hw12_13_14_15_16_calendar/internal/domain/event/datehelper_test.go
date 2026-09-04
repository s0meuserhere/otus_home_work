package event_test

import (
	"errors"
	"testing"
	"time"

	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/event"
)

func TestIsFirstDay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		date    time.Time
		check   func(time.Time) error
		wantErr error
	}{
		{
			name:  "month first day",
			date:  time.Date(2030, 3, 1, 12, 0, 0, 0, time.UTC),
			check: event.IsFirstDayOfMonth,
		},
		{
			name:    "month not first day",
			date:    time.Date(2030, 3, 2, 12, 0, 0, 0, time.UTC),
			check:   event.IsFirstDayOfMonth,
			wantErr: event.ErrFirstDayOfMonth,
		},
		{
			name:  "week monday",
			date:  time.Date(2030, 6, 3, 0, 0, 0, 0, time.UTC),
			check: event.IsFirstDayOfWeek,
		},
		{
			name:    "week tuesday",
			date:    time.Date(2030, 6, 4, 0, 0, 0, 0, time.UTC),
			check:   event.IsFirstDayOfWeek,
			wantErr: event.ErrFirstDayOfWeek,
		},
		{
			name:    "week sunday",
			date:    time.Date(2030, 6, 2, 0, 0, 0, 0, time.UTC),
			check:   event.IsFirstDayOfWeek,
			wantErr: event.ErrFirstDayOfWeek,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.check(tt.date)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err: got %v, want %v", err, tt.wantErr)
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
		})
	}
}
