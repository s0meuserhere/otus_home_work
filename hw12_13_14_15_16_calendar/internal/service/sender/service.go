package sender

import (
	"context"
	"fmt"
	"time"

	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/notify"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/logger"
)

// Service — контракт сервиса рассыльщика.
type Service interface {
	// Send отправляет уведомление пользователю.
	Send(ctx context.Context, n notify.Notify) error
}

type StatusRepository interface {
	SaveStatus(ctx context.Context, n notify.Notify, status notify.Status, at time.Time) error
}

type service struct {
	statuses StatusRepository
}

// New создаёт сервис рассыльщика.
func New(statuses StatusRepository) Service {
	return &service{statuses: statuses}
}

// Send пишет уведомление в лог вместо реальной отправки и сохраняет статус.
func (s *service) Send(ctx context.Context, n notify.Notify) error {
	logger.FromContext(ctx).Info("notify sent",
		"event_id", n.GetEventID(),
		"event_title", n.GetEventTitle(),
		"event_date", n.GetEventDateStart(),
		"user_id", n.GetUserID(),
	)

	if err := s.statuses.SaveStatus(ctx, n, notify.StatusSent, time.Now().UTC()); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}
