package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/domain/event"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/server/http/gen"
)

func (s *Server) GetHello(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("hello"))
}

func (s *Server) CreateEvent(w http.ResponseWriter, r *http.Request) {
	req, err := decodeEventRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	id, title, start, end, description, userID, notifyShift, err := parseEventFields(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	created, err := s.events.Create(r.Context(), id, title, start, end, description, userID, notifyShift)
	if err != nil {
		writeDomainError(w, err)

		return
	}

	writeJSON(w, http.StatusCreated, toOpenAPIEvent(*created))
}

func (s *Server) UpdateEvent(w http.ResponseWriter, r *http.Request, id gen.EventID) {
	req, err := decodeEventRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	_, title, start, end, description, userID, notifyShift, err := parseEventFields(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	updated, err := s.events.Update(r.Context(), id, title, start, end, description, userID, notifyShift)
	if err != nil {
		writeDomainError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, toOpenAPIEvent(*updated))
}

func (s *Server) DeleteEvent(w http.ResponseWriter, r *http.Request, id gen.EventID) {
	if err := s.events.Delete(r.Context(), id); err != nil {
		writeDomainError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) ListEventsByDay(w http.ResponseWriter, r *http.Request, params gen.ListEventsByDayParams) {
	s.handleList(w, r, params.Date.Time, s.events.ListByDay)
}

func (s *Server) ListEventsByWeek(w http.ResponseWriter, r *http.Request, params gen.ListEventsByWeekParams) {
	s.handleList(w, r, params.Date.Time, s.events.ListByWeek)
}

func (s *Server) ListEventsByMonth(w http.ResponseWriter, r *http.Request, params gen.ListEventsByMonthParams) {
	s.handleList(w, r, params.Date.Time, s.events.ListByMonth)
}

func (s *Server) handleList(
	w http.ResponseWriter,
	r *http.Request,
	date time.Time,
	listFn func(context.Context, time.Time) ([]event.Event, error),
) {
	list, err := listFn(r.Context(), date)
	if err != nil {
		writeDomainError(w, err)

		return
	}

	resp := gen.EventList{Events: make([]gen.Event, 0, len(list))}
	for i := range list {
		resp.Events = append(resp.Events, toOpenAPIEvent(list[i]))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) serveOpenAPISpec(w http.ResponseWriter, _ *http.Request) {
	swagger, err := gen.GetSwagger()
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("load openapi: %w", err))

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	data, err := swagger.MarshalJSON()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}
	_, _ = w.Write(data)
}

func decodeEventRequest(r *http.Request) (gen.EventRequest, error) {
	defer r.Body.Close()

	var req gen.EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return gen.EventRequest{}, fmt.Errorf("decode body: %w", err)
	}

	return req, nil
}

func parseEventFields(
	req gen.EventRequest,
) (uuid.UUID, string, time.Time, time.Time, string, uuid.UUID, int, error) {
	id := uuid.Nil
	if req.Id != nil {
		id = *req.Id
	}

	if req.DateStart.IsZero() {
		return uuid.Nil, "", time.Time{}, time.Time{}, "", uuid.Nil, 0, fmt.Errorf("invalid date_start")
	}

	if req.DateEnd.IsZero() {
		return uuid.Nil, "", time.Time{}, time.Time{}, "", uuid.Nil, 0, fmt.Errorf("invalid date_end")
	}

	notifyShift := 0
	if req.NotifyShiftSeconds != nil {
		notifyShift = *req.NotifyShiftSeconds
	}

	return id, req.Title, req.DateStart, req.DateEnd, req.Description, req.UserId, notifyShift, nil
}

func toOpenAPIEvent(e event.Event) gen.Event {
	return gen.Event{
		Id:                 e.GetID(),
		Title:              e.GetTitle(),
		DateStart:          e.GetDateStart().UTC(),
		DateEnd:            e.GetDateEnd().UTC(),
		Description:        e.GetDescription(),
		UserId:             e.GetUserID(),
		NotifyShiftSeconds: e.GetNotifyShiftSeconds(),
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, gen.Error{Error: err.Error()})
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, event.ErrValidation),
		errors.Is(err, event.ErrFirstDayOfWeek),
		errors.Is(err, event.ErrFirstDayOfMonth):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, event.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, event.ErrDateBusy), errors.Is(err, event.ErrAlreadyExists):
		writeError(w, http.StatusConflict, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}
