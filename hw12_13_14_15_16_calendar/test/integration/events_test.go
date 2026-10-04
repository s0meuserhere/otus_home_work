//go:build integration

package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/server/http/gen"
)

func TestCreateEvent(t *testing.T) {
	start := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	req := newEvent("create", start, time.Hour, 600)

	created := createEvent(t, req)

	if created.Title != req.Title || created.UserId != req.UserId || !created.DateStart.Equal(req.DateStart) {
		t.Fatalf("unexpected event: %+v", created)
	}
	if !hasEvent(listEvents(t, "day", start.Format(time.DateOnly)), created.Id) {
		t.Fatal("created event is not listed for its day")
	}
}

func TestUpdateAndDeleteEvent(t *testing.T) {
	start := time.Now().Add(48 * time.Hour).Truncate(time.Second)
	created := createEvent(t, newEvent("before update", start, time.Hour, 0))

	update := newEvent("after update", start.Add(2*time.Hour), time.Hour, 0)
	update.UserId = created.UserId

	var updated gen.Event
	if code := request(t, http.MethodPut, "/events/"+created.Id.String(), update, &updated); code != http.StatusOK {
		t.Fatalf("update: status %d", code)
	}
	if updated.Title != "after update" {
		t.Fatalf("title: got %q", updated.Title)
	}

	if code := request(t, http.MethodDelete, "/events/"+created.Id.String(), nil, nil); code != http.StatusNoContent {
		t.Fatalf("delete: status %d", code)
	}
	if hasEvent(listEvents(t, "day", start.Format(time.DateOnly)), created.Id) {
		t.Fatal("deleted event is still listed")
	}
}

func TestEventErrors(t *testing.T) {
	start := time.Now().Add(72 * time.Hour).Truncate(time.Second)
	existing := createEvent(t, newEvent("existing", start, time.Hour, 0))

	// То же время у того же пользователя.
	busy := newEvent("busy", start.Add(30*time.Minute), time.Hour, 0)
	busy.UserId = existing.UserId

	duplicate := newEvent("duplicate", start.Add(24*time.Hour), time.Hour, 0)
	duplicate.Id = &existing.Id

	noDescription := newEvent("no description", start, time.Hour, 0)
	noDescription.Description = ""

	past := newEvent("past", time.Now().Add(-time.Hour), time.Minute, 0)
	missing := "/events/" + uuid.NewString()

	tests := []struct {
		name   string
		method string
		path   string
		body   any
		want   int
	}{
		{"empty title", http.MethodPost, "/events", newEvent("", start, time.Hour, 0), http.StatusBadRequest},
		{"empty description", http.MethodPost, "/events", noDescription, http.StatusBadRequest},
		{"date in past", http.MethodPost, "/events", past, http.StatusBadRequest},
		{"end before start", http.MethodPost, "/events", newEvent("end", start, -time.Hour, 0), http.StatusBadRequest},
		{"date is busy", http.MethodPost, "/events", busy, http.StatusConflict},
		{"duplicate id", http.MethodPost, "/events", duplicate, http.StatusConflict},
		{"update missing", http.MethodPut, missing, newEvent("missing", start, time.Hour, 0), http.StatusNotFound},
		{"delete missing", http.MethodDelete, missing, nil, http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp gen.Error
			if code := request(t, tt.method, tt.path, tt.body, &resp); code != tt.want {
				t.Fatalf("status: got %d, want %d (%s)", code, tt.want, resp.Error)
			}
			if resp.Error == "" {
				t.Fatal("expected error message")
			}
		})
	}
}
