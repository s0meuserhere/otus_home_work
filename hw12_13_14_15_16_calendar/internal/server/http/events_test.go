package internalhttp_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	eventrepo "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/repository/event"
	internalhttp "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/server/http"
	eventservice "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/service/event"
)

type eventResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type listResponse struct {
	Events []eventResponse `json:"events"`
}

func newHandler(t *testing.T) http.Handler {
	t.Helper()

	logg := logger.New("error", "local")
	svc := eventservice.New(eventrepo.NewMemory())

	return internalhttp.NewServer(logg, ":0", svc).Handler()
}

func eventBody(t *testing.T, id, userID uuid.UUID, title string, start time.Time) []byte {
	t.Helper()

	body := map[string]any{
		"title":                title,
		"date_start":           start.Format(time.RFC3339),
		"date_end":             start.Add(time.Hour).Format(time.RFC3339),
		"description":          "desc",
		"user_id":              userID.String(),
		"notify_shift_seconds": 0,
	}
	if id != uuid.Nil {
		body["id"] = id.String()
	}

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	return raw
}

func do(t *testing.T, h http.Handler, method, target string, body []byte) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return rec
}

func mustCreate(t *testing.T, h http.Handler, body []byte) eventResponse {
	t.Helper()

	rec := do(t, h, http.MethodPost, "/events", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status: got %d body %s", rec.Code, rec.Body.String())
	}

	var created eventResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create: %v", err)
	}

	return created
}

func mustList(t *testing.T, h http.Handler, target string) listResponse {
	t.Helper()

	rec := do(t, h, http.MethodGet, target, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list %s status: got %d body %s", target, rec.Code, rec.Body.String())
	}

	var list listResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}

	return list
}

func TestHTTPEvents(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	// 2030-06-17 - понедельник, 2030-06-01 - первое число месяца.
	start := time.Date(2030, 6, 18, 10, 0, 0, 0, time.UTC)

	t.Run("create and list by day", func(t *testing.T) {
		t.Parallel()

		h := newHandler(t)
		mustCreate(t, h, eventBody(t, uuid.Nil, userID, "meeting", start))

		if got := len(mustList(t, h, "/events/day?date=2030-06-18").Events); got != 1 {
			t.Fatalf("day: expected 1 event, got %d", got)
		}
		if got := len(mustList(t, h, "/events/day?date=2030-06-19").Events); got != 0 {
			t.Fatalf("other day: expected 0 events, got %d", got)
		}
	})

	t.Run("list by week and month", func(t *testing.T) {
		t.Parallel()

		h := newHandler(t)
		mustCreate(t, h, eventBody(t, uuid.Nil, userID, "this week", start))
		mustCreate(t, h, eventBody(t, uuid.Nil, userID, "next week", start.AddDate(0, 0, 7)))
		mustCreate(t, h, eventBody(t, uuid.Nil, userID, "next month", start.AddDate(0, 1, 0)))

		if got := len(mustList(t, h, "/events/week?date=2030-06-17").Events); got != 1 {
			t.Fatalf("week: expected 1 event, got %d", got)
		}
		if got := len(mustList(t, h, "/events/month?date=2030-06-01").Events); got != 2 {
			t.Fatalf("month: expected 2 events, got %d", got)
		}
	})

	t.Run("list with invalid period start", func(t *testing.T) {
		t.Parallel()

		h := newHandler(t)
		for _, target := range []string{"/events/week?date=2030-06-18", "/events/month?date=2030-06-02"} {
			if rec := do(t, h, http.MethodGet, target, nil); rec.Code != http.StatusBadRequest {
				t.Fatalf("%s status: got %d, want 400", target, rec.Code)
			}
		}
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()

		h := newHandler(t)
		created := mustCreate(t, h, eventBody(t, uuid.Nil, userID, "meeting", start))

		rec := do(t, h, http.MethodPut, "/events/"+created.ID,
			eventBody(t, uuid.Nil, userID, "updated", start.Add(2*time.Hour)))
		if rec.Code != http.StatusOK {
			t.Fatalf("update status: got %d body %s", rec.Code, rec.Body.String())
		}

		var updated eventResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
			t.Fatalf("unmarshal update: %v", err)
		}
		if updated.ID != created.ID || updated.Title != "updated" {
			t.Fatalf("unexpected event: %+v", updated)
		}

		list := mustList(t, h, "/events/day?date=2030-06-18")
		if len(list.Events) != 1 || list.Events[0].Title != "updated" {
			t.Fatalf("unexpected list after update: %+v", list.Events)
		}
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()

		h := newHandler(t)
		created := mustCreate(t, h, eventBody(t, uuid.Nil, userID, "meeting", start))

		if rec := do(t, h, http.MethodDelete, "/events/"+created.ID, nil); rec.Code != http.StatusNoContent {
			t.Fatalf("delete status: got %d body %s", rec.Code, rec.Body.String())
		}
		if got := len(mustList(t, h, "/events/day?date=2030-06-18").Events); got != 0 {
			t.Fatalf("expected 0 events after delete, got %d", got)
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		h := newHandler(t)
		missing := "/events/" + uuid.NewString()

		if rec := do(t, h, http.MethodDelete, missing, nil); rec.Code != http.StatusNotFound {
			t.Fatalf("delete status: got %d, want 404", rec.Code)
		}

		rec := do(t, h, http.MethodPut, missing, eventBody(t, uuid.Nil, userID, "meeting", start))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("update status: got %d, want 404", rec.Code)
		}
	})

	t.Run("conflict", func(t *testing.T) {
		t.Parallel()

		h := newHandler(t)
		created := mustCreate(t, h, eventBody(t, uuid.Nil, userID, "meeting", start))

		// Пересечение по времени у того же пользователя.
		rec := do(t, h, http.MethodPost, "/events",
			eventBody(t, uuid.Nil, userID, "overlap", start.Add(30*time.Minute)))
		if rec.Code != http.StatusConflict {
			t.Fatalf("busy status: got %d, want 409", rec.Code)
		}

		// Повтор существующего id.
		rec = do(t, h, http.MethodPost, "/events",
			eventBody(t, uuid.MustParse(created.ID), uuid.New(), "duplicate", start.AddDate(0, 0, 1)))
		if rec.Code != http.StatusConflict {
			t.Fatalf("duplicate status: got %d, want 409", rec.Code)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		h := newHandler(t)
		rec := do(t, h, http.MethodPost, "/events", eventBody(t, uuid.Nil, userID, "", start))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want 400", rec.Code)
		}
	})

	t.Run("body too large", func(t *testing.T) {
		t.Parallel()

		h := newHandler(t)
		body := []byte(`{"title":"` + strings.Repeat("a", 2*1024*1024) + `"}`)
		rec := do(t, h, http.MethodPost, "/events", body)
		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("status: got %d, want 413", rec.Code)
		}
	})
}
