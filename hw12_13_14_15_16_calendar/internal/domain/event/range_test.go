package event_test

import (
	"errors"
	"testing"
	"time"

	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/event"
)

func TestRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		date     time.Time
		build    func(time.Time) (*event.Range, error)
		wantFrom time.Time
		wantTo   time.Time
		wantErr  error
	}{
		{
			name: "day",
			date: time.Date(2030, 6, 15, 14, 30, 0, 0, time.UTC),
			build: func(d time.Time) (*event.Range, error) {
				return event.DayRange(d), nil
			},
			wantFrom: time.Date(2030, 6, 15, 0, 0, 0, 0, time.UTC),
			wantTo:   time.Date(2030, 6, 16, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "week monday",
			date:     time.Date(2030, 6, 3, 10, 0, 0, 0, time.UTC),
			build:    event.WeekRange,
			wantFrom: time.Date(2030, 6, 3, 0, 0, 0, 0, time.UTC),
			wantTo:   time.Date(2030, 6, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "week not monday",
			date:    time.Date(2030, 6, 4, 0, 0, 0, 0, time.UTC),
			build:   event.WeekRange,
			wantErr: event.ErrFirstDayOfWeek,
		},
		{
			name:     "month first day",
			date:     time.Date(2030, 6, 1, 12, 0, 0, 0, time.UTC),
			build:    event.MonthRange,
			wantFrom: time.Date(2030, 6, 1, 0, 0, 0, 0, time.UTC),
			wantTo:   time.Date(2030, 7, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "month not first day",
			date:    time.Date(2030, 6, 2, 0, 0, 0, 0, time.UTC),
			build:   event.MonthRange,
			wantErr: event.ErrFirstDayOfMonth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.build(tt.date)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err: got %v, want %v", err, tt.wantErr)
				}
				if got != nil {
					t.Fatal("expected nil range")
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got == nil {
				t.Fatal("expected range")
			}
			if !got.From.Equal(tt.wantFrom) {
				t.Fatalf("from: got %v, want %v", got.From, tt.wantFrom)
			}
			if !got.To.Equal(tt.wantTo) {
				t.Fatalf("to: got %v, want %v", got.To, tt.wantTo)
			}
		})
	}
}
