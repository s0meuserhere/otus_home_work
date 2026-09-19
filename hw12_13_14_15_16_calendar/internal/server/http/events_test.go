package internalhttp_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	eventrepo "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/repository/event"
	internalhttp "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/server/http"
	eventservice "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/service/event"
)

func TestHTTPEvents(t *testing.T) {
	t.Parallel()

	logg := logger.New("error", "local")
	svc := eventservice.New(eventrepo.NewMemory())
	server := internalhttp.NewServer(logg, ":0", svc)
	handler := server.Handler()

	userID := uuid.New()
	day := time.Date(2030, 6, 15, 0, 0, 0, 0, time.UTC)
	start := day.Add(10 * time.Hour)
	end := start.Add(time.Hour)

	t.Run("create and list by day", func(t *testing.T) {
		t.Parallel()

		body := map[string]any{
			"title":                "meeting",
			"date_start":           start.Format(time.RFC3339),
			"date_end":             end.Format(time.RFC3339),
			"description":          "desc",
			"user_id":              userID.String(),
			"notify_shift_seconds": 0,
		}
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(raw))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("create status: got %d body %s", rec.Code, rec.Body.String())
		}

		listReq := httptest.NewRequest(http.MethodGet, "/events/day?date=2030-06-15", nil)
		listRec := httptest.NewRecorder()
		handler.ServeHTTP(listRec, listReq)

		if listRec.Code != http.StatusOK {
			t.Fatalf("list status: got %d body %s", listRec.Code, listRec.Body.String())
		}

		var listResp struct {
			Events []map[string]any `json:"events"`
		}
		if err := json.Unmarshal(listRec.Body.Bytes(), &listResp); err != nil {
			t.Fatalf("unmarshal list: %v", err)
		}
		if len(listResp.Events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(listResp.Events))
		}
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		body := map[string]any{
			"title":       "",
			"date_start":  start.Format(time.RFC3339),
			"date_end":    end.Format(time.RFC3339),
			"description": "desc",
			"user_id":     userID.String(),
		}
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(raw))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want 400", rec.Code)
		}
	})
}
