package event_test

import (
	"testing"
	"time"

	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/event"
)

func TestTimeRangesIntersect(t *testing.T) {
	t.Parallel()

	base := time.Date(2030, 1, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		aStart time.Time
		aEnd   time.Time
		bStart time.Time
		bEnd   time.Time
		want   bool
	}{
		{
			name:   "intersect",
			aStart: base,
			aEnd:   base.Add(2 * time.Hour),
			bStart: base.Add(time.Hour),
			bEnd:   base.Add(3 * time.Hour),
			want:   true,
		},
		{
			name:   "end equals next start",
			aStart: base,
			aEnd:   base.Add(time.Hour),
			bStart: base.Add(time.Hour),
			bEnd:   base.Add(2 * time.Hour),
			want:   false,
		},
		{
			name:   "no intersect",
			aStart: base,
			aEnd:   base.Add(time.Hour),
			bStart: base.Add(2 * time.Hour),
			bEnd:   base.Add(3 * time.Hour),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := event.TimeRangesIntersect(tt.aStart, tt.aEnd, tt.bStart, tt.bEnd)
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
