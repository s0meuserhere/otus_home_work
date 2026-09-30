package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/notify"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/logger"
)

// Service — контракт сервиса планировщика.
type Service interface {
	// SendPending отправляет в очередь уведомления, для которых наступило время.
	SendPending(ctx context.Context, now time.Time) error
	// DeleteOld удаляет события, закончившиеся раньше чем retention назад.
	DeleteOld(ctx context.Context, now time.Time) error
}

type EventRepository interface {
	DeleteEndedBefore(ctx context.Context, before time.Time) (int64, error)
}

type NotifyRepository interface {
	ListPending(ctx context.Context, now time.Time) ([]notify.Notify, error)
	MarkSent(ctx context.Context, eventID uuid.UUID, at time.Time) error
}

type NotifyQueue interface {
	Publish(ctx context.Context, n notify.Notify) error
}

type service struct {
	events    EventRepository
	notifies  NotifyRepository
	queue     NotifyQueue
	retention time.Duration
}

// New создаёт сервис планировщика.
func New(events EventRepository, notifies NotifyRepository, queue NotifyQueue, retention time.Duration) Service {
	return &service{
		events:    events,
		notifies:  notifies,
		queue:     queue,
		retention: retention,
	}
}

// SendPending отмечает уведомление отправленным только после публикации и возвращает все сбои одной ошибкой.
func (s *service) SendPending(ctx context.Context, now time.Time) error {
	list, err := s.notifies.ListPending(ctx, now)
	if err != nil {
		return fmt.Errorf("send pending: %w", err)
	}

	var errs []error
	for i := range list {
		n := list[i]
		if err := s.send(ctx, n, now); err != nil {
			errs = append(errs, fmt.Errorf("event %s: %w", n.GetEventID(), err))
		}
	}

	logger.FromContext(ctx).Info("pending notifies sent", "found", len(list), "sent", len(list)-len(errs))

	if len(errs) > 0 {
		return fmt.Errorf("send pending: %w", errors.Join(errs...))
	}

	return nil
}

// send публикует уведомление и затем отмечает его отправленным. Доставка at-least-once.
// Если после публикации MarkSent не сработает, уведомление уйдёт повторно при следующем запуске.
func (s *service) send(ctx context.Context, n notify.Notify, now time.Time) error {
	if err := s.queue.Publish(ctx, n); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	if err := s.notifies.MarkSent(ctx, n.GetEventID(), now); err != nil {
		return fmt.Errorf("mark sent: %w", err)
	}

	return nil
}

// DeleteOld удаляет события, закончившиеся раньше чем retention назад.
func (s *service) DeleteOld(ctx context.Context, now time.Time) error {
	before := now.Add(-s.retention)

	deleted, err := s.events.DeleteEndedBefore(ctx, before)
	if err != nil {
		return fmt.Errorf("delete old events ended before %s: %w", before.Format(time.RFC3339), err)
	}

	if deleted > 0 {
		logger.FromContext(ctx).Info("old events deleted", "count", deleted, "ended_before", before)
	}

	return nil
}
