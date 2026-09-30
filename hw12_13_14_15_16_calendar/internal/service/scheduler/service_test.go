package scheduler_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/notify"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/service/scheduler"
)

const retention = 365 * 24 * time.Hour

type fakeEvents struct {
	mu            sync.Mutex
	deletedBefore []time.Time
	err           error
}

func (r *fakeEvents) DeleteEndedBefore(_ context.Context, before time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.err != nil {
		return 0, r.err
	}
	r.deletedBefore = append(r.deletedBefore, before)

	return 1, nil
}

type fakeNotifies struct {
	mu       sync.Mutex
	pending  []notify.Notify
	listErr  error
	listedAt []time.Time
	marked   []uuid.UUID
	markErr  error
}

func (r *fakeNotifies) ListPending(_ context.Context, now time.Time) ([]notify.Notify, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.listedAt = append(r.listedAt, now)

	return r.pending, r.listErr
}

func (r *fakeNotifies) MarkSent(_ context.Context, eventID uuid.UUID, _ time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.markErr != nil {
		return r.markErr
	}
	r.marked = append(r.marked, eventID)

	return nil
}

type fakeQueue struct {
	mu     sync.Mutex
	failOn map[uuid.UUID]bool
	sent   []uuid.UUID
}

func (q *fakeQueue) Publish(_ context.Context, n notify.Notify) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.failOn[n.GetEventID()] {
		return errors.New("broker unavailable")
	}
	q.sent = append(q.sent, n.GetEventID())

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

func TestSendPendingPublishesAndMarks(t *testing.T) {
	t.Parallel()

	now := time.Date(2030, 6, 15, 12, 0, 0, 0, time.UTC)
	n1, n2 := newNotify(t), newNotify(t)
	notifies := &fakeNotifies{pending: []notify.Notify{n1, n2}}
	queue := &fakeQueue{}
	svc := scheduler.New(&fakeEvents{}, notifies, queue, retention)

	if err := svc.SendPending(context.Background(), now); err != nil {
		t.Fatalf("send pending: %v", err)
	}

	if len(notifies.listedAt) != 1 || !notifies.listedAt[0].Equal(now) {
		t.Fatalf("list pending: got %v, want one call with %v", notifies.listedAt, now)
	}
	if len(queue.sent) != 2 || queue.sent[0] != n1.GetEventID() || queue.sent[1] != n2.GetEventID() {
		t.Fatalf("published: got %v", queue.sent)
	}
	if len(notifies.marked) != 2 || notifies.marked[0] != n1.GetEventID() || notifies.marked[1] != n2.GetEventID() {
		t.Fatalf("marked: got %v", notifies.marked)
	}
}

func TestSendPendingDoesNotMarkFailedPublish(t *testing.T) {
	t.Parallel()

	ok, failed := newNotify(t), newNotify(t)
	notifies := &fakeNotifies{pending: []notify.Notify{failed, ok}}
	queue := &fakeQueue{failOn: map[uuid.UUID]bool{failed.GetEventID(): true}}
	svc := scheduler.New(&fakeEvents{}, notifies, queue, retention)

	// Сбой одного уведомления возвращается, но не мешает остальным.
	err := svc.SendPending(context.Background(), time.Now())
	if err == nil || !strings.Contains(err.Error(), failed.GetEventID().String()) {
		t.Fatalf("err: got %v, want error mentioning %s", err, failed.GetEventID())
	}

	// Неотправленное не отмечается и уйдёт при следующем запуске.
	if len(notifies.marked) != 1 || notifies.marked[0] != ok.GetEventID() {
		t.Fatalf("marked: got %v, want only %s", notifies.marked, ok.GetEventID())
	}
}

func TestSendPendingMarkError(t *testing.T) {
	t.Parallel()

	n := newNotify(t)
	notifies := &fakeNotifies{pending: []notify.Notify{n}, markErr: errors.New("db is down")}
	queue := &fakeQueue{}
	svc := scheduler.New(&fakeEvents{}, notifies, queue, retention)

	// Уведомление опубликовано, но не отмечено, поэтому уйдёт повторно (at-least-once).
	if err := svc.SendPending(context.Background(), time.Now()); err == nil {
		t.Fatal("expected error")
	}
	if len(queue.sent) != 1 {
		t.Fatalf("published: got %d, want 1", len(queue.sent))
	}
}

func TestSendPendingListError(t *testing.T) {
	t.Parallel()

	notifies := &fakeNotifies{listErr: errors.New("db is down")}
	queue := &fakeQueue{}
	svc := scheduler.New(&fakeEvents{}, notifies, queue, retention)

	if err := svc.SendPending(context.Background(), time.Now()); err == nil {
		t.Fatal("expected error")
	}
	if len(queue.sent) != 0 {
		t.Fatalf("published: got %d, want 0", len(queue.sent))
	}
}

func TestDeleteOld(t *testing.T) {
	t.Parallel()

	now := time.Date(2030, 6, 15, 12, 0, 0, 0, time.UTC)
	events := &fakeEvents{}
	svc := scheduler.New(events, &fakeNotifies{}, &fakeQueue{}, retention)

	if err := svc.DeleteOld(context.Background(), now); err != nil {
		t.Fatalf("delete old: %v", err)
	}

	want := now.Add(-retention)
	if len(events.deletedBefore) != 1 || !events.deletedBefore[0].Equal(want) {
		t.Fatalf("deleted before: got %v, want %v", events.deletedBefore, want)
	}
}

func TestDeleteOldError(t *testing.T) {
	t.Parallel()

	events := &fakeEvents{err: errors.New("db is down")}
	svc := scheduler.New(events, &fakeNotifies{}, &fakeQueue{}, retention)

	if err := svc.DeleteOld(context.Background(), time.Now()); err == nil {
		t.Fatal("expected error")
	}
}
