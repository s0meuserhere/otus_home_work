package sender_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/notify"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/service/sender"
)

type fakeStatuses struct {
	saved  []notify.Status
	events []uuid.UUID
	err    error
}

func (r *fakeStatuses) SaveStatus(_ context.Context, n notify.Notify, status notify.Status, _ time.Time) error {
	if r.err != nil {
		return r.err
	}
	r.saved = append(r.saved, status)
	r.events = append(r.events, n.GetEventID())

	return nil
}

func newNotify(t *testing.T) notify.Notify {
	t.Helper()

	n, err := notify.NewNotify(uuid.New(), "meeting", time.Now().Add(time.Hour), uuid.New())
	if err != nil {
		t.Fatalf("new notify: %v", err)
	}

	return *n
}

func TestSendSavesStatus(t *testing.T) {
	t.Parallel()

	n := newNotify(t)
	statuses := &fakeStatuses{}

	if err := sender.New(statuses).Send(context.Background(), n); err != nil {
		t.Fatalf("send: %v", err)
	}

	if len(statuses.saved) != 1 || statuses.saved[0] != notify.StatusSent || statuses.events[0] != n.GetEventID() {
		t.Fatalf("saved: got %v for %v", statuses.saved, statuses.events)
	}
}

func TestSendStatusError(t *testing.T) {
	t.Parallel()

	statuses := &fakeStatuses{err: errors.New("db is down")}

	if err := sender.New(statuses).Send(context.Background(), newNotify(t)); err == nil {
		t.Fatal("expected error")
	}
}
