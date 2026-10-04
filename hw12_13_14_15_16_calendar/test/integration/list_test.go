//go:build integration

package integration

import (
	"net/http"
	"testing"
	"time"
)

// 1 февраля 2100 года - понедельник, поэтому эта дата подходит как начало дня, недели и месяца.
// База в тестовом окружении каждый раз новая, других событий в этом месяце нет.
var month = time.Date(2100, 2, 1, 0, 0, 0, 0, time.UTC)

func TestListEvents(t *testing.T) {
	day1 := createEvent(t, newEvent("day 1", month.Add(10*time.Hour), time.Hour, 0))
	day3 := createEvent(t, newEvent("day 3", month.AddDate(0, 0, 2).Add(10*time.Hour), time.Hour, 0))
	day8 := createEvent(t, newEvent("day 8", month.AddDate(0, 0, 7).Add(10*time.Hour), time.Hour, 0))
	createEvent(t, newEvent("next month", month.AddDate(0, 1, 1).Add(10*time.Hour), time.Hour, 0))

	date := month.Format(time.DateOnly)

	if events := listEvents(t, "day", date); len(events) != 1 || !hasEvent(events, day1.Id) {
		t.Fatalf("day: got %d events", len(events))
	}

	if events := listEvents(t, "week", date); len(events) != 2 || !hasEvent(events, day3.Id) {
		t.Fatalf("week: got %d events", len(events))
	}

	if events := listEvents(t, "month", date); len(events) != 3 || !hasEvent(events, day8.Id) {
		t.Fatalf("month: got %d events", len(events))
	}
}

func TestListEventsWrongPeriodStart(t *testing.T) {
	tuesday := month.AddDate(0, 0, 1).Format(time.DateOnly)

	if code := request(t, http.MethodGet, "/events/week?date="+tuesday, nil, nil); code != http.StatusBadRequest {
		t.Fatalf("week from tuesday: status %d", code)
	}

	if code := request(t, http.MethodGet, "/events/month?date="+tuesday, nil, nil); code != http.StatusBadRequest {
		t.Fatalf("month from 2nd day: status %d", code)
	}
}
