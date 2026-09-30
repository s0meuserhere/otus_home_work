package notify

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/event"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/notify"
)

type notifyRow struct {
	EventID    uuid.UUID
	EventTitle string
	DateStart  time.Time
	UserID     uuid.UUID
}

func (r notifyRow) toDomain() (*notify.Notify, error) {
	return notify.NewNotify(r.EventID, r.EventTitle, r.DateStart, r.UserID)
}

// DB берёт уведомления из таблицы events и отмечает отправку в notified_at.
type DB struct {
	pool *pgxpool.Pool
}

func NewDB(pool *pgxpool.Pool) *DB {
	return &DB{pool: pool}
}

// ListPending возвращает неотправленные уведомления, время которых наступило.
// Опоздавшие после простоя уведомления тоже уходят, но только пока событие не закончилось.
func (d *DB) ListPending(ctx context.Context, now time.Time) ([]notify.Notify, error) {
	const query = `
		SELECT id, title, date_start, user_id
		FROM events
		WHERE notified_at IS NULL
			AND notify_shift_seconds > 0
			AND date_start - make_interval(secs => notify_shift_seconds) <= $1
			AND date_end > $1
		ORDER BY date_start`

	rows, err := d.pool.Query(ctx, query, now)
	if err != nil {
		return nil, fmt.Errorf("list pending notifies: %w", err)
	}
	defer rows.Close()

	result := make([]notify.Notify, 0)

	for rows.Next() {
		var row notifyRow
		if err := rows.Scan(&row.EventID, &row.EventTitle, &row.DateStart, &row.UserID); err != nil {
			return nil, fmt.Errorf("scan notify: %w", err)
		}

		n, err := row.toDomain()
		if err != nil {
			return nil, fmt.Errorf("convert notify: %w", err)
		}

		result = append(result, *n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifies: %w", err)
	}

	return result, nil
}

// MarkSent отмечает, что уведомление по событию отправлено.
func (d *DB) MarkSent(ctx context.Context, eventID uuid.UUID, at time.Time) error {
	const query = `UPDATE events SET notified_at = $2 WHERE id = $1`

	tag, err := d.pool.Exec(ctx, query, eventID, at)
	if err != nil {
		return fmt.Errorf("mark notify sent: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", event.ErrNotFound, eventID)
	}

	return nil
}
