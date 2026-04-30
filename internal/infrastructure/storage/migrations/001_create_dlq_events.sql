CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS dlq_events (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id       TEXT        NOT NULL DEFAULT '',
    topic          TEXT        NOT NULL,
    payload        JSONB       NOT NULL,
    error          TEXT        NOT NULL,
    source_service TEXT        NOT NULL,
    retry_count    INT         NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dlq_events_source_service ON dlq_events (source_service);
CREATE INDEX IF NOT EXISTS idx_dlq_events_created_at     ON dlq_events (created_at DESC);
