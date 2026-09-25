package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/event"
)

type eventRow struct {
	ID                 uuid.UUID
	Title              string
	DateStart          time.Time
	DateEnd            time.Time
	Description        string
	UserID             uuid.UUID
	NotifyShiftSeconds int
}

func (r eventRow) toDomain() (*event.Event, error) {
	return event.NewEvent(r.ID, r.Title, r.DateStart, r.DateEnd, r.Description, r.UserID, r.NotifyShiftSeconds)
}

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(pool *pgxpool.Pool) *DB {
	return &DB{pool: pool}
}

func (d *DB) Create(ctx context.Context, e event.Event) error {
	const query = `
		INSERT INTO events (
			id, title, date_start, date_end, description, user_id, notify_shift_seconds
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := d.pool.Exec(
		ctx,
		query,
		e.GetID(),
		e.GetTitle(),
		e.GetDateStart(),
		e.GetDateEnd(),
		e.GetDescription(),
		e.GetUserID(),
		e.GetNotifyShiftSeconds(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%w: %s", event.ErrAlreadyExists, e.GetID())
		}

		return fmt.Errorf("insert event: %w", err)
	}

	return nil
}

func (d *DB) Update(ctx context.Context, eventID uuid.UUID, e event.Event) error {
	const query = `
		UPDATE events
		SET title = $2,
			date_start = $3,
			date_end = $4,
			description = $5,
			user_id = $6,
			notify_shift_seconds = $7
		WHERE id = $1`

	tag, err := d.pool.Exec(
		ctx,
		query,
		eventID,
		e.GetTitle(),
		e.GetDateStart(),
		e.GetDateEnd(),
		e.GetDescription(),
		e.GetUserID(),
		e.GetNotifyShiftSeconds(),
	)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", event.ErrNotFound, eventID)
	}

	return nil
}

func (d *DB) Delete(ctx context.Context, eventID uuid.UUID) error {
	const query = `DELETE FROM events WHERE id = $1`

	tag, err := d.pool.Exec(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", event.ErrNotFound, eventID)
	}

	return nil
}

func (d *DB) ListInRange(ctx context.Context, r *event.Range) ([]event.Event, error) {
	const query = `
		SELECT id, title, date_start, date_end, description, user_id, notify_shift_seconds
		FROM events
		WHERE date_start >= $1 AND date_start < $2
		ORDER BY date_start`

	rows, err := d.pool.Query(ctx, query, r.From, r.To)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	result := make([]event.Event, 0)

	for rows.Next() {
		var row eventRow
		if err := rows.Scan(
			&row.ID,
			&row.Title,
			&row.DateStart,
			&row.DateEnd,
			&row.Description,
			&row.UserID,
			&row.NotifyShiftSeconds,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		ev, err := row.toDomain()
		if err != nil {
			return nil, fmt.Errorf("convert event: %w", err)
		}

		result = append(result, *ev)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	return result, nil
}

func (d *DB) IsDateBusy(
	ctx context.Context,
	userID, excludeID uuid.UUID,
	start, end time.Time,
) (bool, error) {
	const query = `
		SELECT id
		FROM events
		WHERE user_id = $1
			AND id <> $2
			AND date_start < $4
			AND date_end > $3
		LIMIT 1`

	var busyID uuid.UUID
	err := d.pool.QueryRow(ctx, query, userID, excludeID, start, end).Scan(&busyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("check date busy: %w", err)
	}

	return true, nil
}
