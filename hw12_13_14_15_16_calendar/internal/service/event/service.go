package event

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/event"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/logger"
)

// Service — контракт сервиса событий.
type Service interface {
	Create(
		ctx context.Context,
		id uuid.UUID,
		title string,
		dateStart time.Time,
		dateEnd time.Time,
		description string,
		userID uuid.UUID,
		notifyShiftSeconds int,
	) (*event.Event, error)
	Update(
		ctx context.Context,
		eventID uuid.UUID,
		title string,
		dateStart time.Time,
		dateEnd time.Time,
		description string,
		userID uuid.UUID,
		notifyShiftSeconds int,
	) (*event.Event, error)
	Delete(ctx context.Context, eventID uuid.UUID) error
	ListByDay(ctx context.Context, date time.Time) ([]event.Event, error)
	ListByWeek(ctx context.Context, date time.Time) ([]event.Event, error)
	ListByMonth(ctx context.Context, date time.Time) ([]event.Event, error)
}

type Repository interface {
	Create(ctx context.Context, e event.Event) error
	Update(ctx context.Context, eventID uuid.UUID, e event.Event) error
	Delete(ctx context.Context, eventID uuid.UUID) error
	ListInRange(ctx context.Context, r *event.Range) ([]event.Event, error)
	IsDateBusy(ctx context.Context, userID, excludeID uuid.UUID, start, end time.Time) (bool, error)
}

type service struct {
	storage Repository
}

// New создаёт сервис событий.
func New(storage Repository) Service {
	return &service{
		storage: storage,
	}
}

// Create создаёт новое событие.
func (s *service) Create(
	ctx context.Context,
	id uuid.UUID,
	title string,
	dateStart time.Time,
	dateEnd time.Time,
	description string,
	userID uuid.UUID,
	notifyShiftSeconds int,
) (*event.Event, error) {
	log := logger.FromContext(ctx)

	e, err := event.NewEvent(id, title, dateStart, dateEnd, description, userID, notifyShiftSeconds)
	if err != nil {
		log.Error("create invalid", "err", err)

		return nil, fmt.Errorf("create: %w", err)
	}

	if err := s.checkDateNotBusy(ctx, e.GetUserID(), e.GetID(), e.GetDateStart(), e.GetDateEnd()); err != nil {
		log.Error("create busy", "id", e.GetID(), "err", err)

		return nil, fmt.Errorf("create: %w", err)
	}

	if err := s.storage.Create(ctx, *e); err != nil {
		log.Error("create failed", "id", e.GetID(), "err", err)

		return nil, fmt.Errorf("create: %w", err)
	}

	log.Info("created", "id", e.GetID(), "user_id", e.GetUserID())

	return e, nil
}

// Update обновляет существующее событие.
func (s *service) Update(
	ctx context.Context,
	eventID uuid.UUID,
	title string,
	dateStart time.Time,
	dateEnd time.Time,
	description string,
	userID uuid.UUID,
	notifyShiftSeconds int,
) (*event.Event, error) {
	log := logger.FromContext(ctx)

	e, err := event.NewEvent(eventID, title, dateStart, dateEnd, description, userID, notifyShiftSeconds)
	if err != nil {
		log.Error("update invalid", "id", eventID, "err", err)

		return nil, fmt.Errorf("update: %w", err)
	}

	if err := s.checkDateNotBusy(ctx, e.GetUserID(), eventID, e.GetDateStart(), e.GetDateEnd()); err != nil {
		log.Error("update busy", "id", eventID, "err", err)

		return nil, fmt.Errorf("update: %w", err)
	}

	if err := s.storage.Update(ctx, eventID, *e); err != nil {
		log.Error("update failed", "id", eventID, "err", err)

		return nil, fmt.Errorf("update: %w", err)
	}

	log.Info("updated", "id", eventID)

	return e, nil
}

// Delete удаляет событие по ID.
func (s *service) Delete(ctx context.Context, eventID uuid.UUID) error {
	log := logger.FromContext(ctx)

	if err := s.storage.Delete(ctx, eventID); err != nil {
		log.Error("delete failed", "id", eventID, "err", err)

		return fmt.Errorf("delete: %w", err)
	}

	log.Info("deleted", "id", eventID)

	return nil
}

// ListByDay возвращает список событий на день.
func (s *service) ListByDay(ctx context.Context, date time.Time) ([]event.Event, error) {
	log := logger.FromContext(ctx)

	list, err := s.storage.ListInRange(ctx, event.DayRange(date))
	if err != nil {
		log.Error("list day failed", "date", date, "err", err)

		return nil, fmt.Errorf("list day: %w", err)
	}

	log.Info("list day", "date", date, "count", len(list))

	return list, nil
}

// ListByWeek возвращает список событий на неделю.
func (s *service) ListByWeek(ctx context.Context, date time.Time) ([]event.Event, error) {
	log := logger.FromContext(ctx)

	r, err := event.WeekRange(date)
	if err != nil {
		log.Error("list week invalid", "date", date, "err", err)

		return nil, fmt.Errorf("list week: %w", err)
	}

	list, err := s.storage.ListInRange(ctx, r)
	if err != nil {
		log.Error("list week failed", "date", date, "err", err)

		return nil, fmt.Errorf("list week: %w", err)
	}

	log.Info("list week", "date", date, "count", len(list))

	return list, nil
}

// ListByMonth возвращает список событий на месяц.
func (s *service) ListByMonth(ctx context.Context, date time.Time) ([]event.Event, error) {
	log := logger.FromContext(ctx)

	r, err := event.MonthRange(date)
	if err != nil {
		log.Error("list month invalid", "date", date, "err", err)

		return nil, fmt.Errorf("list month: %w", err)
	}

	list, err := s.storage.ListInRange(ctx, r)
	if err != nil {
		log.Error("list month failed", "date", date, "err", err)

		return nil, fmt.Errorf("list month: %w", err)
	}

	log.Info("list month", "date", date, "count", len(list))

	return list, nil
}

func (s *service) checkDateNotBusy(
	ctx context.Context,
	userID, excludeID uuid.UUID,
	start, end time.Time,
) error {
	busy, err := s.storage.IsDateBusy(ctx, userID, excludeID, start, end)
	if err != nil {
		return fmt.Errorf("date busy check: %w", err)
	}

	if busy {
		return fmt.Errorf("%w: user %s", event.ErrDateBusy, userID)
	}

	return nil
}
