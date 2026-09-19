package internalgrpc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/event"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/server/grpc/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateEvent(ctx context.Context, req *gen.CreateEventRequest) (*gen.CreateEventResponse, error) {
	id, title, start, end, description, userID, notifyShift, err := parseCreateUpdate(
		req.GetId(),
		req.GetTitle(),
		req.GetDateStart(),
		req.GetDateEnd(),
		req.GetDescription(),
		req.GetUserId(),
		req.GetNotifyShiftSeconds(),
	)
	if err != nil {
		return nil, toGRPCError(err)
	}

	created, err := s.events.Create(ctx, id, title, start, end, description, userID, notifyShift)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &gen.CreateEventResponse{Event: toProtoEvent(*created)}, nil
}

func (s *Server) UpdateEvent(ctx context.Context, req *gen.UpdateEventRequest) (*gen.UpdateEventResponse, error) {
	eventID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	_, title, start, end, description, userID, notifyShift, err := parseCreateUpdate(
		"",
		req.GetTitle(),
		req.GetDateStart(),
		req.GetDateEnd(),
		req.GetDescription(),
		req.GetUserId(),
		req.GetNotifyShiftSeconds(),
	)
	if err != nil {
		return nil, toGRPCError(err)
	}

	updated, err := s.events.Update(ctx, eventID, title, start, end, description, userID, notifyShift)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &gen.UpdateEventResponse{Event: toProtoEvent(*updated)}, nil
}

func (s *Server) DeleteEvent(ctx context.Context, req *gen.DeleteEventRequest) (*gen.DeleteEventResponse, error) {
	eventID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	if err := s.events.Delete(ctx, eventID); err != nil {
		return nil, toGRPCError(err)
	}

	return &gen.DeleteEventResponse{}, nil
}

func (s *Server) ListEventsByDay(ctx context.Context, req *gen.ListEventsRequest) (*gen.ListEventsResponse, error) {
	return s.list(ctx, req, s.events.ListByDay)
}

func (s *Server) ListEventsByWeek(ctx context.Context, req *gen.ListEventsRequest) (*gen.ListEventsResponse, error) {
	return s.list(ctx, req, s.events.ListByWeek)
}

func (s *Server) ListEventsByMonth(ctx context.Context, req *gen.ListEventsRequest) (*gen.ListEventsResponse, error) {
	return s.list(ctx, req, s.events.ListByMonth)
}

func (s *Server) list(
	ctx context.Context,
	req *gen.ListEventsRequest,
	listFn func(context.Context, time.Time) ([]event.Event, error),
) (*gen.ListEventsResponse, error) {
	if req.GetDate() == nil {
		return nil, status.Error(codes.InvalidArgument, "date is required")
	}

	list, err := listFn(ctx, req.GetDate().AsTime())
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := &gen.ListEventsResponse{Events: make([]*gen.Event, 0, len(list))}
	for i := range list {
		resp.Events = append(resp.Events, toProtoEvent(list[i]))
	}

	return resp, nil
}

func parseCreateUpdate(
	idStr, title string,
	dateStart, dateEnd *timestamppb.Timestamp,
	description, userIDStr string,
	notifyShift int32,
) (uuid.UUID, string, time.Time, time.Time, string, uuid.UUID, int, error) {
	var id uuid.UUID
	var err error

	if idStr != "" {
		id, err = uuid.Parse(idStr)
		if err != nil {
			return uuid.Nil, "", time.Time{}, time.Time{}, "", uuid.Nil, 0,
				fmt.Errorf("%w: invalid id", event.ErrValidation)
		}
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, "", time.Time{}, time.Time{}, "", uuid.Nil, 0,
			fmt.Errorf("%w: invalid user_id", event.ErrValidation)
	}

	if dateStart == nil {
		return uuid.Nil, "", time.Time{}, time.Time{}, "", uuid.Nil, 0,
			fmt.Errorf("%w: invalid date_start", event.ErrValidation)
	}

	if dateEnd == nil {
		return uuid.Nil, "", time.Time{}, time.Time{}, "", uuid.Nil, 0,
			fmt.Errorf("%w: invalid date_end", event.ErrValidation)
	}

	return id, title, dateStart.AsTime(), dateEnd.AsTime(), description, userID, int(notifyShift), nil
}

func toProtoEvent(e event.Event) *gen.Event {
	notifyShift := e.GetNotifyShiftSeconds()
	if notifyShift < 0 {
		notifyShift = 0
	}
	if notifyShift > math.MaxInt32 {
		notifyShift = math.MaxInt32
	}

	return &gen.Event{
		Id:                 e.GetID().String(),
		Title:              e.GetTitle(),
		DateStart:          timestamppb.New(e.GetDateStart().UTC()),
		DateEnd:            timestamppb.New(e.GetDateEnd().UTC()),
		Description:        e.GetDescription(),
		UserId:             e.GetUserID().String(),
		NotifyShiftSeconds: int32(notifyShift), //nolint:gosec // value is clamped to MaxInt32
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, event.ErrValidation),
		errors.Is(err, event.ErrFirstDayOfWeek),
		errors.Is(err, event.ErrFirstDayOfMonth):
		return status.Errorf(codes.InvalidArgument, "%v", err)
	case errors.Is(err, event.ErrNotFound):
		return status.Errorf(codes.NotFound, "%v", err)
	case errors.Is(err, event.ErrDateBusy), errors.Is(err, event.ErrAlreadyExists):
		return status.Errorf(codes.AlreadyExists, "%v", err)
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
