package sender

import (
	"context"

	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/notify"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/logger"
)

// Service — контракт сервиса рассыльщика.
type Service interface {
	// Send отправляет уведомление пользователю.
	Send(ctx context.Context, n notify.Notify) error
}

type service struct{}

// New создаёт сервис рассыльщика.
func New() Service {
	return &service{}
}

// Send пишет уведомление в лог вместо реальной отправки.
func (s *service) Send(ctx context.Context, n notify.Notify) error {
	logger.FromContext(ctx).Info("notify sent",
		"event_id", n.GetEventID(),
		"event_title", n.GetEventTitle(),
		"event_date", n.GetEventDateStart(),
		"user_id", n.GetUserID(),
	)

	return nil
}
