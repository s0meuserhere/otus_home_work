-- +goose Up
ALTER TABLE events ADD COLUMN IF NOT EXISTS notified_at TIMESTAMPTZ NULL;

-- индекс под выборку планировщика
CREATE INDEX IF NOT EXISTS events_pending_notify_idx ON events (date_start)
    WHERE notified_at IS NULL AND notify_shift_seconds > 0;

CREATE INDEX IF NOT EXISTS events_date_end_idx ON events (date_end);

-- +goose Down
DROP INDEX IF EXISTS events_date_end_idx;
DROP INDEX IF EXISTS events_pending_notify_idx;
ALTER TABLE events DROP COLUMN IF EXISTS notified_at;
