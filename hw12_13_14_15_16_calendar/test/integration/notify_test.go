//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// waitStatus ждёт до 30 секунд, пока рассыльщик запишет статус уведомления.
func waitStatus(t *testing.T, eventID uuid.UUID) string {
	t.Helper()

	for i := 0; i < 30; i++ {
		var status string
		// Ошибку не проверяем, пока строки нет, Scan вернёт pgx.ErrNoRows.
		_ = db.QueryRow(context.Background(),
			`SELECT status FROM notify_statuses WHERE event_id = $1`, eventID).Scan(&status)
		if status != "" {
			return status
		}
		time.Sleep(time.Second)
	}

	return ""
}

func TestNotifySent(t *testing.T) {
	// Событие через 10 минут с напоминанием за час, значит напоминать уже пора.
	event := createEvent(t, newEvent("remind me", time.Now().Add(10*time.Minute), 2*time.Hour, 3600))

	if status := waitStatus(t, event.Id); status != "sent" {
		t.Fatalf("notify status: got %q, want sent", status)
	}
}

func TestNotifyNotSentTooEarly(t *testing.T) {
	noReminder := createEvent(t, newEvent("no reminder", time.Now().Add(10*time.Minute), time.Hour, 0))
	later := createEvent(t, newEvent("remind later", time.Now().Add(72*time.Hour), 2*time.Hour, 3600))
	due := createEvent(t, newEvent("remind now", time.Now().Add(10*time.Minute), 2*time.Hour, 3600))

	// Когда дошло уведомление, которому пора, планировщик уже видел и остальные события.
	if status := waitStatus(t, due.Id); status != "sent" {
		t.Fatalf("due event: got %q, want sent", status)
	}

	for _, id := range []uuid.UUID{noReminder.Id, later.Id} {
		var count int
		if err := db.QueryRow(context.Background(),
			`SELECT count(*) FROM notify_statuses WHERE event_id = $1`, id).Scan(&count); err != nil {
			t.Fatalf("count statuses: %v", err)
		}
		if count != 0 {
			t.Fatalf("event %s must not be notified yet", id)
		}
	}
}
