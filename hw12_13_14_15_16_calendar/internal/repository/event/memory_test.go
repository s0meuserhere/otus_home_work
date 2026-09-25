package event

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/event"
)

func testEvent(t *testing.T, userID uuid.UUID, start, end time.Time) event.Event {
	t.Helper()

	e, err := event.NewEvent(
		uuid.New(),
		"title",
		start,
		end,
		"description",
		userID,
		0,
	)
	if err != nil {
		t.Fatalf("new event: %v", err)
	}

	return *e
}

func TestMemory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "create and list in range",
			run: func(t *testing.T) {
				t.Helper()
				repo := NewMemory()
				ctx := context.Background()
				userID := uuid.New()

				day := time.Date(2030, 6, 15, 0, 0, 0, 0, time.UTC)
				e1 := testEvent(t, userID, day.Add(10*time.Hour), day.Add(11*time.Hour))
				e2 := testEvent(t, userID, day.AddDate(0, 0, 1).Add(10*time.Hour), day.AddDate(0, 0, 1).Add(11*time.Hour))

				if err := repo.Create(ctx, e1); err != nil {
					t.Fatalf("create e1: %v", err)
				}
				if err := repo.Create(ctx, e2); err != nil {
					t.Fatalf("create e2: %v", err)
				}

				list, err := repo.ListInRange(ctx, event.DayRange(day))
				if err != nil {
					t.Fatalf("list in range: %v", err)
				}
				if len(list) != 1 {
					t.Fatalf("expected 1 event, got %d", len(list))
				}
				if list[0].GetID() != e1.GetID() {
					t.Fatalf("unexpected event id: %s", list[0].GetID())
				}
			},
		},
		{
			name: "create duplicate",
			run: func(t *testing.T) {
				t.Helper()
				repo := NewMemory()
				ctx := context.Background()
				userID := uuid.New()
				start := time.Date(2030, 7, 1, 10, 0, 0, 0, time.UTC)
				e := testEvent(t, userID, start, start.Add(time.Hour))

				if err := repo.Create(ctx, e); err != nil {
					t.Fatalf("create: %v", err)
				}

				err := repo.Create(ctx, e)
				if !errors.Is(err, event.ErrAlreadyExists) {
					t.Fatalf("expected ErrAlreadyExists, got %v", err)
				}
			},
		},
		{
			name: "is date busy",
			run: func(t *testing.T) {
				t.Helper()
				repo := NewMemory()
				ctx := context.Background()
				userID := uuid.New()
				start := time.Date(2030, 8, 1, 10, 0, 0, 0, time.UTC)

				e1 := testEvent(t, userID, start, start.Add(2*time.Hour))
				if err := repo.Create(ctx, e1); err != nil {
					t.Fatalf("create e1: %v", err)
				}

				busy, err := repo.IsDateBusy(ctx, userID, uuid.New(), start.Add(time.Hour), start.Add(3*time.Hour))
				if err != nil {
					t.Fatalf("is date busy: %v", err)
				}
				if !busy {
					t.Fatal("expected date busy")
				}

				busy, err = repo.IsDateBusy(ctx, userID, e1.GetID(), start, start.Add(2*time.Hour))
				if err != nil {
					t.Fatalf("is date busy exclude self: %v", err)
				}
				if busy {
					t.Fatal("expected date free when excluding self")
				}
			},
		},
		{
			name: "update and delete",
			run: func(t *testing.T) {
				t.Helper()
				repo := NewMemory()
				ctx := context.Background()
				userID := uuid.New()
				start := time.Date(2030, 9, 1, 10, 0, 0, 0, time.UTC)
				e := testEvent(t, userID, start, start.Add(time.Hour))

				if err := repo.Create(ctx, e); err != nil {
					t.Fatalf("create: %v", err)
				}

				updatedPtr, err := event.NewEvent(
					e.GetID(),
					"updated",
					start.Add(2*time.Hour),
					start.Add(3*time.Hour),
					"new description",
					userID,
					60,
				)
				if err != nil {
					t.Fatalf("new updated event: %v", err)
				}

				if err := repo.Update(ctx, e.GetID(), *updatedPtr); err != nil {
					t.Fatalf("update: %v", err)
				}

				list, err := repo.ListInRange(ctx, event.DayRange(start))
				if err != nil {
					t.Fatalf("list: %v", err)
				}
				if len(list) != 1 {
					t.Fatalf("expected 1 event, got %d", len(list))
				}
				if list[0].GetTitle() != "updated" {
					t.Fatalf("expected updated title, got %s", list[0].GetTitle())
				}

				if err := repo.Delete(ctx, e.GetID()); err != nil {
					t.Fatalf("delete: %v", err)
				}

				err = repo.Delete(ctx, e.GetID())
				if !errors.Is(err, event.ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "concurrent create",
			run: func(t *testing.T) {
				t.Helper()
				repo := NewMemory()
				ctx := context.Background()
				userID := uuid.New()

				const n = 50
				var wg sync.WaitGroup
				wg.Add(n)

				errCh := make(chan error, n)

				for i := 0; i < n; i++ {
					go func(i int) {
						defer wg.Done()

						start := time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * 2 * time.Hour)
						e := testEvent(t, userID, start, start.Add(time.Hour))
						errCh <- repo.Create(ctx, e)
					}(i)
				}

				wg.Wait()
				close(errCh)

				for err := range errCh {
					if err != nil {
						t.Fatalf("concurrent create: %v", err)
					}
				}

				month, err := event.MonthRange(time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC))
				if err != nil {
					t.Fatalf("month range: %v", err)
				}

				list, err := repo.ListInRange(ctx, month)
				if err != nil {
					t.Fatalf("list in range: %v", err)
				}
				if len(list) != n {
					t.Fatalf("expected %d events, got %d", n, len(list))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.run(t)
		})
	}
}
