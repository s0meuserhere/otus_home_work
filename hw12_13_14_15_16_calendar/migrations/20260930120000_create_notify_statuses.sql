-- +goose Up
-- Рассыльщик записывает сюда результат обработки каждого уведомления.
CREATE TABLE IF NOT EXISTS notify_statuses (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS notify_statuses_event_id_idx ON notify_statuses (event_id);

-- +goose Down
DROP INDEX IF EXISTS notify_statuses_event_id_idx;
DROP TABLE IF EXISTS notify_statuses;
