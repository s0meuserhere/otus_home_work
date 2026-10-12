//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/server/http/gen"
)

var (
	baseURL string
	db      *pgxpool.Pool
)

// Compose запускает тесты, когда календарь уже прошёл healthcheck, поэтому ждать готовности не нужно.
func TestMain(m *testing.M) {
	baseURL = os.Getenv("CALENDAR_HTTP_URL")

	pool, err := pgxpool.New(context.Background(), os.Getenv("CALENDAR_DB_DSN"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	db = pool

	code := m.Run()
	pool.Close()
	os.Exit(code)
}

// request отправляет запрос в API и возвращает код ответа. Если out не nil, в него читается тело ответа.
func request(t *testing.T, method, path string, body, out any) int {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, baseURL+path, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("%s %s: decode response: %v", method, path, err)
		}
	}

	return resp.StatusCode
}

// newEvent возвращает событие нового пользователя. shift - за сколько секунд напомнить.
func newEvent(title string, start time.Time, duration time.Duration, shift int) gen.EventRequest {
	return gen.EventRequest{
		Title:              title,
		DateStart:          start.UTC(),
		DateEnd:            start.Add(duration).UTC(),
		Description:        "integration test",
		UserId:             uuid.New(),
		NotifyShiftSeconds: &shift,
	}
}

func createEvent(t *testing.T, req gen.EventRequest) gen.Event {
	t.Helper()

	var created gen.Event
	if code := request(t, http.MethodPost, "/events", req, &created); code != http.StatusCreated {
		t.Fatalf("create event: status %d", code)
	}

	return created
}

func listEvents(t *testing.T, period, date string) []gen.Event {
	t.Helper()

	var list gen.EventList
	if code := request(t, http.MethodGet, "/events/"+period+"?date="+date, nil, &list); code != http.StatusOK {
		t.Fatalf("list %s %s: status %d", period, date, code)
	}

	return list.Events
}

func hasEvent(events []gen.Event, id uuid.UUID) bool {
	for _, e := range events {
		if e.Id == id {
			return true
		}
	}

	return false
}
