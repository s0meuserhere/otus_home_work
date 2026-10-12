package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/notify"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/infrastructure"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/logger"
)

// notifyMessage - формат уведомления в очереди.
type notifyMessage struct {
	EventID    uuid.UUID `json:"eventId"`
	EventTitle string    `json:"eventTitle"`
	EventDate  time.Time `json:"eventDate"`
	UserID     uuid.UUID `json:"userId"`
}

func toMessage(n notify.Notify) ([]byte, error) {
	data, err := json.Marshal(notifyMessage{
		EventID:    n.GetEventID(),
		EventTitle: n.GetEventTitle(),
		EventDate:  n.GetEventDateStart().UTC(),
		UserID:     n.GetUserID(),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal notify: %w", err)
	}

	return data, nil
}

func fromMessage(data []byte) (*notify.Notify, error) {
	var m notifyMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%w: unmarshal notify: %w", notify.ErrValidation, err)
	}

	return notify.NewNotify(m.EventID, m.EventTitle, m.EventDate, m.UserID)
}

// Rabbit передаёт уведомления через очередь RabbitMQ.
type Rabbit struct {
	client *infrastructure.RabbitClient
}

func NewRabbit(client *infrastructure.RabbitClient) *Rabbit {
	return &Rabbit{client: client}
}

// Publish кладёт уведомление в очередь и ждёт подтверждения от брокера.
func (r *Rabbit) Publish(ctx context.Context, n notify.Notify) error {
	body, err := toMessage(n)
	if err != nil {
		return err
	}

	if err := r.client.Publish(ctx, body); err != nil {
		return fmt.Errorf("publish notify: %w", err)
	}

	return nil
}

// Consume передаёт уведомления в handle, а нераспознанные пишет в лог и отклоняет в очередь недоставленных.
func (r *Rabbit) Consume(
	ctx context.Context,
	prefetch int,
	handle func(ctx context.Context, n notify.Notify) error,
) error {
	err := r.client.Consume(ctx, prefetch, func(ctx context.Context, body []byte) error {
		n, err := fromMessage(body)
		if err != nil {
			logger.FromContext(ctx).Error("invalid notify message", "body", string(body), "err", err)

			return err
		}

		if err := handle(ctx, *n); err != nil {
			logger.FromContext(ctx).Error("handle notify failed", "event_id", n.GetEventID(), "err", err)

			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("consume notifies: %w", err)
	}

	return nil
}
