-- +goose Up
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    date_start TIMESTAMPTZ NOT NULL,
    date_end TIMESTAMPTZ NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    user_id UUID NOT NULL,
    notify_shift_seconds INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT events_date_range_check CHECK (date_end > date_start)
);

CREATE INDEX IF NOT EXISTS events_user_date_start_idx ON events (user_id, date_start);
CREATE INDEX IF NOT EXISTS events_date_start_idx ON events (date_start);

-- +goose Down
DROP INDEX IF EXISTS events_date_start_idx;
DROP INDEX IF EXISTS events_user_date_start_idx;
DROP TABLE IF EXISTS events;
